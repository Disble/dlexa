package cache

import (
	"context"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestSearchMemoryStoreStoresNormalizedSearchResults(t *testing.T) {
	store := NewSearchMemoryStore()
	result := model.SearchResult{Candidates: []model.SearchCandidate{{RawLabelHTML: `<em>Abu Dhabi</em>`, DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi"}}}
	if err := store.Set(context.Background(), BuildSearchKey(model.SearchRequest{Query: " abu   dhabi ", Format: "json"}), result); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	cached, ok, err := store.Get(context.Background(), BuildSearchKey(model.SearchRequest{Query: "abu dhabi", Format: "markdown"}))
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if len(cached.Candidates) != 1 || cached.Candidates[0].DisplayText != "Abu Dhabi" {
		t.Fatalf("cached = %#v", cached)
	}
}
