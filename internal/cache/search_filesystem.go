package cache

import (
	"context"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

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
	return readFilesystemEntry[model.SearchResult](s.dir, s.now, key)
}

// Set writes the search result to a JSON file with an expiry envelope.
func (s *SearchFilesystemStore) Set(ctx context.Context, key string, result model.SearchResult) error {
	_ = ctx
	return writeFilesystemEntry(s.dir, s.ttl, s.now, key, result, "dlexa-search-cache-*.tmp")
}

func (s *SearchFilesystemStore) pathForKey(key string) string {
	return cachePath(s.dir, key)
}
