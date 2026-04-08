package cache

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
)

// SearchMemoryStore is an in-memory cache for normalized search results.
type SearchMemoryStore struct {
	items memoryItems[model.SearchResult]
}

// NewSearchMemoryStore returns a ready-to-use in-memory SearchStore.
func NewSearchMemoryStore() *SearchMemoryStore {
	return &SearchMemoryStore{items: newMemoryItems[model.SearchResult]()}
}

// Get retrieves a cached search result by key.
func (s *SearchMemoryStore) Get(ctx context.Context, key string) (model.SearchResult, bool, error) {
	_ = ctx
	result, ok := s.items.get(key)
	return result, ok, nil
}

// Set stores a search result under the given key.
func (s *SearchMemoryStore) Set(ctx context.Context, key string, result model.SearchResult) error {
	_ = ctx
	s.items.set(key, result)
	return nil
}
