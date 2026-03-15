package cache

import (
	"context"
	"sync"
	"testing"
)

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
		go runConcurrentWriter(t, store, w, iterations, &wg)
	}

	// Reader goroutines: each reads from the shared key and random unique keys.
	for r := 0; r < readers; r++ {
		go runConcurrentReader(t, store, r, iterations, writers, &wg)
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
