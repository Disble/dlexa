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

// cacheEnvelope wraps a LookupResult with TTL metadata for filesystem persistence.
type cacheEnvelope struct {
	ExpiresAt time.Time          `json:"expires_at"`
	CreatedAt time.Time          `json:"created_at"`
	Data      model.LookupResult `json:"data"`
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
	var zero model.LookupResult

	path := s.pathForKey(key)

	data, err := os.ReadFile(path) //nolint:gosec // G304: path is derived from SHA256 hash of key, not user input
	if err != nil {
		return zero, false, nil
	}

	var envelope cacheEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		// Corrupted file: remove it silently.
		_ = os.Remove(path)
		return zero, false, nil
	}

	// TTL check: if expired, treat as miss.
	if s.now().After(envelope.ExpiresAt) {
		_ = os.Remove(path)
		return zero, false, nil
	}

	return envelope.Data, true, nil
}

// Set writes the result to a JSON file with an expiry envelope.
// Uses atomic write (temp file + rename). On any error, returns nil (silent failure).
func (s *FilesystemStore) Set(ctx context.Context, key string, result model.LookupResult) error {
	_ = ctx

	// Ensure directory exists.
	if err := os.MkdirAll(s.dir, 0750); err != nil {
		return nil // graceful degradation: silent failure
	}

	now := s.now()
	envelope := cacheEnvelope{
		ExpiresAt: now.Add(s.ttl),
		CreatedAt: now,
		Data:      result,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return nil // graceful degradation
	}

	path := s.pathForKey(key)

	// Atomic write: write to temp file, then rename.
	tmpFile, err := os.CreateTemp(s.dir, "dlexa-cache-*.tmp")
	if err != nil {
		return nil // graceful degradation
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return nil // graceful degradation
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return nil // graceful degradation
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return nil // graceful degradation
	}

	return nil
}

// pathForKey returns the full filesystem path for a cache key.
// The key is hashed with SHA256 and hex-encoded to produce a safe filename.
func (s *FilesystemStore) pathForKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	filename := fmt.Sprintf("%x.json", hash)
	return filepath.Join(s.dir, filename)
}
