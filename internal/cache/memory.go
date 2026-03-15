package cache

import (
	"context"
	"strings"
	"sync"

	"github.com/gentleman-programming/dlexa/internal/model"
)

// MemoryStore is an in-memory cache backed by a sync.RWMutex-protected map.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]model.LookupResult
}

// NewMemoryStore returns a ready-to-use in-memory Store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: map[string]model.LookupResult{}}
}

// Get retrieves a cached result by key, delegating to GetResult.
func (s *MemoryStore) Get(ctx context.Context, key string) (model.LookupResult, bool, error) {
	return s.GetResult(ctx, key)
}

// Set stores a result under the given key, delegating to SetResult.
func (s *MemoryStore) Set(ctx context.Context, key string, result model.LookupResult) error {
	return s.SetResult(ctx, key, result)
}

// GetResult retrieves a cached LookupResult, returning false if not found.
func (s *MemoryStore) GetResult(ctx context.Context, key string) (model.LookupResult, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.items[key]
	return result, ok, nil
}

// SetResult stores a LookupResult under the given key.
func (s *MemoryStore) SetResult(ctx context.Context, key string, result model.LookupResult) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = result
	return nil
}

// BuildKey produces a deterministic cache key from a LookupRequest.
func BuildKey(request model.LookupRequest) string {
	parts := []string{
		strings.TrimSpace(request.Query),
		strings.TrimSpace(request.Format),
		strings.Join(request.Sources, ","),
	}

	return strings.Join(parts, "|")
}
