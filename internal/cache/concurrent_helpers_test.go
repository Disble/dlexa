package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func runConcurrentWriter(t *testing.T, store Store, id, iterations int, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()
	ctx := context.Background()
	for i := 0; i < iterations; i++ {
		sharedResult := model.LookupResult{
			Request: model.LookupRequest{Query: "shared"},
			Entries: []model.Entry{{ID: fmt.Sprintf("writer-%d-iter-%d", id, i)}},
		}
		if err := store.Set(ctx, keyShared, sharedResult); err != nil {
			t.Errorf("Set(shared) error = %v", err)
			return
		}

		uniqueKey := fmt.Sprintf("key-%d-%d", id, i)
		uniqueResult := model.LookupResult{
			Request: model.LookupRequest{Query: uniqueKey},
			Entries: []model.Entry{{ID: uniqueKey}},
		}
		if err := store.Set(ctx, uniqueKey, uniqueResult); err != nil {
			t.Errorf("Set(unique) error = %v", err)
			return
		}
	}
}

func runConcurrentReader(t *testing.T, store Store, id, iterations, writers int, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()
	ctx := context.Background()
	for i := 0; i < iterations; i++ {
		if _, _, err := store.Get(ctx, keyShared); err != nil {
			t.Errorf("Get(shared) error = %v", err)
			return
		}

		otherKey := fmt.Sprintf("key-%d-%d", id%writers, i)
		if _, _, err := store.Get(ctx, otherKey); err != nil {
			t.Errorf("Get(other) error = %v", err)
			return
		}
	}
}
