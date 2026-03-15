package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func runMemoryWriter(t *testing.T, store *MemoryStore, id, iterations int, wg *sync.WaitGroup) {
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

func runMemoryReader(t *testing.T, store *MemoryStore, id, iterations, writers int, wg *sync.WaitGroup) {
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

func TestMemoryStoreConcurrentReadWrite(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	const writers = 10
	const readers = 10
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	// Writer goroutines: each writes to a mix of shared and unique keys.
	for w := 0; w < writers; w++ {
		go runMemoryWriter(t, store, w, iterations, &wg)
	}

	// Reader goroutines: each reads from the shared key and random unique keys.
	for r := 0; r < readers; r++ {
		go runMemoryReader(t, store, r, iterations, writers, &wg)
	}

	wg.Wait()

	// Verify at least the shared key was written.
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
