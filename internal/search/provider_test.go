package search

import (
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/model"
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
