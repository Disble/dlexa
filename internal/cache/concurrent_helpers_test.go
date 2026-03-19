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

func runConcurrentStoreTest(t *testing.T, store Store, writers, readers, iterations int) {
	t.Helper()
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for w := 0; w < writers; w++ {
		go runConcurrentWriter(t, store, w, iterations, &wg)
	}
	for r := 0; r < readers; r++ {
		go runConcurrentReader(t, store, r, iterations, writers, &wg)
	}

	wg.Wait()

	result, ok, err := store.Get(ctx, keyShared)
	if err != nil {
		t.Fatalf("final Get(shared) error = %v", err)
	}
	if !ok {
		t.Fatal("final Get(shared) ok = false, want true")
	}
	if len(result.Entries) == 0 {
		t.Fatal("final Get(shared) entries = 0, want > 0")
	}
}
