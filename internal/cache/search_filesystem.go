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

type searchCacheEnvelope struct {
	ExpiresAt time.Time          `json:"expires_at"`
	CreatedAt time.Time          `json:"created_at"`
	Data      model.SearchResult `json:"data"`
}

// SearchFilesystemStore persists search cache entries as JSON files.
type SearchFilesystemStore struct {
	dir string
	ttl time.Duration
	now func() time.Time
}

// NewSearchFilesystemStore creates a store that persists normalized search entries under dir.
func NewSearchFilesystemStore(dir string, ttl time.Duration) *SearchFilesystemStore {
	return &SearchFilesystemStore{dir: dir, ttl: ttl, now: time.Now}
}

// Get returns the cached search result if present and not expired.
func (s *SearchFilesystemStore) Get(ctx context.Context, key string) (model.SearchResult, bool, error) {
	_ = ctx
	var zero model.SearchResult

	path := s.pathForKey(key)
	data, err := os.ReadFile(path) //nolint:gosec // G304: path derived from hashed cache key
	if err != nil {
		return zero, false, nil
	}

	var envelope searchCacheEnvelope
	if json.Unmarshal(data, &envelope) != nil {
		_ = os.Remove(path)
		return zero, false, nil
	}

	if s.now().After(envelope.ExpiresAt) {
		_ = os.Remove(path)
		return zero, false, nil
	}

	return envelope.Data, true, nil
}

// Set writes the search result to a JSON file with an expiry envelope.
func (s *SearchFilesystemStore) Set(ctx context.Context, key string, result model.SearchResult) error {
	_ = ctx
	if os.MkdirAll(s.dir, 0o750) != nil {
		return nil
	}

	now := s.now()
	envelope := searchCacheEnvelope{ExpiresAt: now.Add(s.ttl), CreatedAt: now, Data: result}
	data, err := json.Marshal(envelope)
	if err != nil {
		return nil
	}

	tmpFile, err := os.CreateTemp(s.dir, "dlexa-search-cache-*.tmp")
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
	if os.Rename(tmpPath, s.pathForKey(key)) != nil {
		_ = os.Remove(tmpPath)
		return nil
	}

	return nil
}

func (s *SearchFilesystemStore) pathForKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return filepath.Join(s.dir, fmt.Sprintf("%x.json", hash))
}
