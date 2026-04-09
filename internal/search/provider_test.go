package search

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
	parseengine "github.com/Disble/dlexa/internal/parse/engine"
)

func TestStaticRegistryUsesConfiguredDefaultProvider(t *testing.T) {
	primary := &providerStub{descriptor: model.SourceDescriptor{Name: "search", Priority: 1}}
	secondary := &providerStub{descriptor: model.SourceDescriptor{Name: "academia", Priority: 2}}
	registry := NewStaticRegistry("academia", primary, secondary)

	providers, err := registry.ProvidersFor(model.SearchRequest{Query: "tilde"})
	if err != nil {
		t.Fatalf("ProvidersFor() error = %v", err)
	}
	if got := providerNames(providers); !reflect.DeepEqual(got, []string{"academia"}) {
		t.Fatalf("provider names = %v, want [academia]", got)
	}
}

func TestStaticRegistryMatchesRequestedProvidersByExactName(t *testing.T) {
	registry := NewStaticRegistry("search",
		&providerStub{descriptor: model.SourceDescriptor{Name: "search", Priority: 1}},
		&providerStub{descriptor: model.SourceDescriptor{Name: "academia", Priority: 2}},
	)

	_, err := registry.ProvidersFor(model.SearchRequest{Query: "tilde", Sources: []string{"SEARCH"}})
	if err == nil {
		t.Fatal("ProvidersFor() error = nil, want exact-name mismatch failure")
	}
}

func providerNames(providers []Provider) []string {
	names := make([]string, 0, len(providers))
	for _, provider := range providers {
		names = append(names, provider.Descriptor().Name)
	}
	return names
}

func TestNewEnginePipelineProviderPreservesLegacySearchParserBehavior(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "search", Priority: 1}
	retrievedAt := time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC)
	legacyParser := &stubParser{records: []parse.ParsedSearchRecord{{Title: "solo", URL: "https://www.rae.es/dpd/solo"}}, warnings: []model.Warning{{Code: "parse-warning", Source: descriptor.Name}}}
	normalizer := &stubNormalizer{candidates: []model.SearchCandidate{{Title: "solo", URL: "https://www.rae.es/dpd/solo"}}, warnings: []model.Warning{{Code: "normalize-warning", Source: descriptor.Name}}}
	provider := NewEnginePipelineProvider(
		descriptor,
		&stubFetcher{document: fetch.Document{URL: "https://example.invalid/search?q=solo", Body: []byte("body"), RetrievedAt: retrievedAt}},
		parseengine.AdaptLegacySearchParser(legacyParser),
		normalizer,
	)
	provider.now = func() time.Time { return retrievedAt }

	result, err := provider.Search(context.Background(), model.SearchRequest{Query: "solo"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].Title != "solo" {
		t.Fatalf("Search() candidates = %#v, want normalized candidates through engine adapter", result.Candidates)
	}
	if !reflect.DeepEqual(legacyParser.document, fetch.Document{URL: "https://example.invalid/search?q=solo", Body: []byte("body"), RetrievedAt: retrievedAt}) {
		t.Fatalf("legacy parser document = %#v, want fetch document propagated", legacyParser.document)
	}
	if len(result.Warnings) != 2 {
		t.Fatalf("Search() warnings len = %d, want parse + normalize warnings", len(result.Warnings))
	}
}
