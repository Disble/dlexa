package cache

import (
	"context"
	"strings"
	"sync"

	"github.com/Disble/dlexa/internal/model"
)

type memoryItems[T any] struct {
	mu    sync.RWMutex
	items map[string]T
}

func newMemoryItems[T any]() memoryItems[T] {
	return memoryItems[T]{items: map[string]T{}}
}

func (m *memoryItems[T]) get(key string) (T, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result, ok := m.items[key]
	return result, ok
}

func (m *memoryItems[T]) set(key string, value T) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[key] = value
}

// MemoryStore is an in-memory cache backed by a sync.RWMutex-protected map.
type MemoryStore struct {
	items memoryItems[model.LookupResult]
}

// NewMemoryStore returns a ready-to-use in-memory Store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: newMemoryItems[model.LookupResult]()}
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
	result, ok := s.items.get(key)
	return result, ok, nil
}

// SetResult stores a LookupResult under the given key.
func (s *MemoryStore) SetResult(ctx context.Context, key string, result model.LookupResult) error {
	_ = ctx
	s.items.set(key, result)
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
