package cache

import (
	"context"
	"strings"
	"sync"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]model.LookupResult
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: map[string]model.LookupResult{}}
}

func (s *MemoryStore) Get(ctx context.Context, key string) (model.LookupResult, bool, error) {
	return s.GetResult(ctx, key)
}

func (s *MemoryStore) Set(ctx context.Context, key string, result model.LookupResult) error {
	return s.SetResult(ctx, key, result)
}

func (s *MemoryStore) GetResult(ctx context.Context, key string) (model.LookupResult, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.items[key]
	return result, ok, nil
}

func (s *MemoryStore) SetResult(ctx context.Context, key string, result model.LookupResult) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = result
	return nil
}

func BuildKey(request model.LookupRequest) string {
	parts := []string{
		strings.TrimSpace(request.Query),
		strings.TrimSpace(request.Format),
		strings.Join(request.Sources, ","),
	}

	return strings.Join(parts, "|")
}
