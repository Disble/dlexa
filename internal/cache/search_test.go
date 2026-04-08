package cache

import (
	"context"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestBuildSearchKeySortsProvidersAndNormalizesQuery(t *testing.T) {
	requestA := model.SearchRequest{Query: " Abu   Dhabi ", Sources: []string{"beta", " alpha ", "beta"}}
	requestB := model.SearchRequest{Query: "Abu Dhabi", Sources: []string{"alpha", "beta"}}

	gotA := BuildSearchKey(requestA)
	gotB := BuildSearchKey(requestB)
	const want = "search/v2|Abu Dhabi|providers=alpha,beta"

	if gotA != want {
		t.Fatalf("BuildSearchKey(requestA) = %q, want %q", gotA, want)
	}
	if gotB != want {
		t.Fatalf("BuildSearchKey(requestB) = %q, want %q", gotB, want)
	}
	if gotA != gotB {
		t.Fatalf("BuildSearchKey() mismatch = %q vs %q", gotA, gotB)
	}
}

func TestBuildSearchKeyKeepsProviderCachesIsolated(t *testing.T) {
	store := NewSearchMemoryStore()
	requestA := model.SearchRequest{Query: "tilde", Sources: []string{"search"}}
	requestB := model.SearchRequest{Query: "tilde", Sources: []string{"academia"}}
	stored := model.SearchResult{Candidates: []model.SearchCandidate{{Title: "solo"}}}

	if err := store.Set(context.Background(), BuildSearchKey(requestA), stored); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if _, ok, err := store.Get(context.Background(), BuildSearchKey(requestA)); err != nil || !ok {
		t.Fatalf("Get(requestA) = (_, %v, %v), want hit", ok, err)
	}
	if _, ok, err := store.Get(context.Background(), BuildSearchKey(requestB)); err != nil {
		t.Fatalf("Get(requestB) error = %v", err)
	} else if ok {
		t.Fatal("Get(requestB) ok = true, want false for provider-isolated key")
	}
}

func TestLegacySearchKeyOnlyAppliesToDefaultSingleProvider(t *testing.T) {
	tests := []struct {
		name            string
		request         model.SearchRequest
		providerName    string
		defaultProvider string
		wantOK          bool
	}{
		{
			name:            "implicit default provider",
			request:         model.SearchRequest{Query: "solo"},
			providerName:    "search",
			defaultProvider: "search",
			wantOK:          true,
		},
		{
			name:            "explicit default provider",
			request:         model.SearchRequest{Query: "solo", Sources: []string{"search"}},
			providerName:    "search",
			defaultProvider: "search",
			wantOK:          true,
		},
		{
			name:            "non-default provider denied",
			request:         model.SearchRequest{Query: "solo", Sources: []string{"academia"}},
			providerName:    "academia",
			defaultProvider: "search",
			wantOK:          false,
		},
		{
			name:            "multi-provider request denied",
			request:         model.SearchRequest{Query: "solo", Sources: []string{"search", "academia"}},
			providerName:    "search",
			defaultProvider: "search",
			wantOK:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, ok := LegacySearchKey(tt.request, tt.providerName, tt.defaultProvider)
			if ok != tt.wantOK {
				t.Fatalf("LegacySearchKey() ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				if key != "" {
					t.Fatalf("LegacySearchKey() key = %q, want empty when disabled", key)
				}
				return
			}
			if want := BuildLegacySearchKey(tt.request); key != want {
				t.Fatalf("LegacySearchKey() = %q, want %q", key, want)
			}
		})
	}
}
