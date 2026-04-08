// Package cache provides caching implementations for dlexa.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

// cacheEnvelope wraps cached data with TTL metadata for filesystem persistence.
type cacheEnvelope[T any] struct {
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Data      T         `json:"data"`
}

func hashedCacheFilename(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x.json", hash)
}

func cachePath(dir, key string) string {
	return filepath.Join(dir, hashedCacheFilename(key))
}

func readFilesystemEntry[T any](dir string, now func() time.Time, key string) (T, bool, error) {
	var zero T
	path := cachePath(dir, key)

	data, err := os.ReadFile(path) //nolint:gosec // G304: path is derived from a SHA256 hash of the cache key
	if err != nil {
		return zero, false, nil
	}

	var envelope cacheEnvelope[T]
	if json.Unmarshal(data, &envelope) != nil {
		_ = os.Remove(path)
		return zero, false, nil
	}
	if now().After(envelope.ExpiresAt) {
		_ = os.Remove(path)
		return zero, false, nil
	}

	return envelope.Data, true, nil
}

func writeFilesystemEntry[T any](dir string, ttl time.Duration, now func() time.Time, key string, result T, tempPattern string) error {
	if os.MkdirAll(dir, 0o750) != nil {
		return nil
	}

	timestamp := now()
	envelope := cacheEnvelope[T]{
		ExpiresAt: timestamp.Add(ttl),
		CreatedAt: timestamp,
		Data:      result,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return nil
	}

	tmpFile, err := os.CreateTemp(dir, tempPattern)
	if err != nil {
		return nil
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return nil
	}
	if tmpFile.Close() != nil {
		_ = os.Remove(tmpPath)
		return nil
	}
	if os.Rename(tmpPath, cachePath(dir, key)) != nil {
		_ = os.Remove(tmpPath)
		return nil
	}

	return nil
}

// FilesystemStore persists cache entries as JSON files under a directory.
// It implements cache.Store with graceful degradation: all filesystem/JSON
// errors are swallowed and treated as cache misses (Get) or no-ops (Set).
type FilesystemStore struct {
	dir string
	ttl time.Duration
	now func() time.Time // injectable clock for testing
}

// NewFilesystemStore creates a store that persists entries under dir.
// dir is created with os.MkdirAll on first write. ttl controls entry lifetime.
func NewFilesystemStore(dir string, ttl time.Duration) *FilesystemStore {
	return &FilesystemStore{
		dir: dir,
		ttl: ttl,
		now: time.Now,
	}
}

// Get returns the cached result if the file exists, is valid JSON, and has not expired.
// On any error (file not found, corrupt JSON, expired), returns (zero, false, nil).
func (s *FilesystemStore) Get(ctx context.Context, key string) (model.LookupResult, bool, error) {
	_ = ctx
	return readFilesystemEntry[model.LookupResult](s.dir, s.now, key)
}

// Set writes the result to a JSON file with an expiry envelope.
// Uses atomic write (temp file + rename). On any error, returns nil (silent failure).
func (s *FilesystemStore) Set(ctx context.Context, key string, result model.LookupResult) error {
	_ = ctx
	return writeFilesystemEntry(s.dir, s.ttl, s.now, key, result, "dlexa-cache-*.tmp")
}

// pathForKey returns the full filesystem path for a cache key.
// The key is hashed with SHA256 and hex-encoded to produce a safe filename.
func (s *FilesystemStore) pathForKey(key string) string {
	return cachePath(s.dir, key)
}
