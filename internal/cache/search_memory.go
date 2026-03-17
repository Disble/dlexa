package cache

import (
	"context"
	"sync"

	"github.com/Disble/dlexa/internal/model"
)

// SearchMemoryStore is an in-memory cache for normalized search results.
type SearchMemoryStore struct {
	mu    sync.RWMutex
	items map[string]model.SearchResult
}

// NewSearchMemoryStore returns a ready-to-use in-memory SearchStore.
func NewSearchMemoryStore() *SearchMemoryStore {
	return &SearchMemoryStore{items: map[string]model.SearchResult{}}
}

// Get retrieves a cached search result by key.
func (s *SearchMemoryStore) Get(ctx context.Context, key string) (model.SearchResult, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.items[key]
	return result, ok, nil
}

// Set stores a search result under the given key.
func (s *SearchMemoryStore) Set(ctx context.Context, key string, result model.SearchResult) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = result
	return nil
}
