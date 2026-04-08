package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

type concurrentStoreHarness[T any] struct {
	get          func(context.Context, string) (T, bool, error)
	set          func(context.Context, string, T) error
	sharedValue  func(id, iteration int) T
	uniqueValue  func(key string) T
	valuePresent func(T) bool
}

func runConcurrentCacheTest[T any](t *testing.T, harness concurrentStoreHarness[T], writers, readers, iterations int) {
	t.Helper()
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for w := 0; w < writers; w++ {
		go runConcurrentWriter(t, harness, w, iterations, &wg)
	}
	for r := 0; r < readers; r++ {
		go runConcurrentReader(t, harness, r, iterations, writers, &wg)
	}

	wg.Wait()

	result, ok, err := harness.get(ctx, keyShared)
	if err != nil {
		t.Fatalf("final Get(shared) error = %v", err)
	}
	if !ok {
		t.Fatal("final Get(shared) ok = false, want true")
	}
	if !harness.valuePresent(result) {
		t.Fatal("final Get(shared) returned empty value")
	}
}

func runConcurrentWriter[T any](t *testing.T, harness concurrentStoreHarness[T], id, iterations int, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()
	ctx := context.Background()
	for i := 0; i < iterations; i++ {
		if err := harness.set(ctx, keyShared, harness.sharedValue(id, i)); err != nil {
			t.Errorf("Set(shared) error = %v", err)
			return
		}

		uniqueKey := fmt.Sprintf("key-%d-%d", id, i)
		if err := harness.set(ctx, uniqueKey, harness.uniqueValue(uniqueKey)); err != nil {
			t.Errorf("Set(unique) error = %v", err)
			return
		}
	}
}

func runConcurrentReader[T any](t *testing.T, harness concurrentStoreHarness[T], id, iterations, writers int, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()
	ctx := context.Background()
	for i := 0; i < iterations; i++ {
		if _, _, err := harness.get(ctx, keyShared); err != nil {
			t.Errorf("Get(shared) error = %v", err)
			return
		}

		otherKey := fmt.Sprintf("key-%d-%d", id%writers, i)
		if _, _, err := harness.get(ctx, otherKey); err != nil {
			t.Errorf("Get(other) error = %v", err)
			return
		}
	}
}

func runConcurrentStoreTest(t *testing.T, store Store, writers, readers, iterations int) {
	t.Helper()
	runConcurrentCacheTest(t, concurrentStoreHarness[model.LookupResult]{
		get: store.Get,
		set: store.Set,
		sharedValue: func(id, iteration int) model.LookupResult {
			return model.LookupResult{
				Request: model.LookupRequest{Query: "shared"},
				Entries: []model.Entry{{ID: fmt.Sprintf("writer-%d-iter-%d", id, iteration)}},
			}
		},
		uniqueValue: func(key string) model.LookupResult {
			return model.LookupResult{
				Request: model.LookupRequest{Query: key},
				Entries: []model.Entry{{ID: key}},
			}
		},
		valuePresent: func(result model.LookupResult) bool { return len(result.Entries) > 0 },
	}, writers, readers, iterations)
}

func runConcurrentSearchStoreTest(t *testing.T, store SearchStore, writers, readers, iterations int) {
	t.Helper()
	runConcurrentCacheTest(t, concurrentStoreHarness[model.SearchResult]{
		get: store.Get,
		set: store.Set,
		sharedValue: func(id, iteration int) model.SearchResult {
			return model.SearchResult{
				Request: model.SearchRequest{Query: "shared"},
				Candidates: []model.SearchCandidate{{
					ArticleKey:  fmt.Sprintf("writer-%d-iter-%d", id, iteration),
					DisplayText: "shared",
				}},
			}
		},
		uniqueValue: func(key string) model.SearchResult {
			return model.SearchResult{
				Request:    model.SearchRequest{Query: key},
				Candidates: []model.SearchCandidate{{ArticleKey: key, DisplayText: key}},
			}
		},
		valuePresent: func(result model.SearchResult) bool { return len(result.Candidates) > 0 },
	}, writers, readers, iterations)
}
