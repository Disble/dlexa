package app

import (
	"testing"

	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/fetch"
	modsearch "github.com/Disble/dlexa/internal/modules/search"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
	searchsvc "github.com/Disble/dlexa/internal/search"
)

func TestNewWiresSearchModuleToLiveSearchAdapters(t *testing.T) {
	application := New(&fakeCLI{})
	module, ok := application.registry.Module("search")
	if !ok {
		t.Fatal("search module not registered")
	}
	searchModule, ok := module.(*modsearch.Module)
	if !ok {
		t.Fatalf("module type = %T, want *search.Module", module)
	}
	searcher := searchModule.SearcherForTesting()
	service, ok := searcher.(*searchsvc.Service)
	if !ok {
		t.Fatalf("searcher type = %T, want *search.Service", searcher)
	}
	registry, ok := service.RegistryForTesting().(*searchsvc.StaticRegistry)
	if !ok {
		t.Fatalf("registry type = %T, want *search.StaticRegistry", service.RegistryForTesting())
	}
	providers := registry.Providers()
	if len(providers) != 2 {
		t.Fatalf("providers len = %d, want 2", len(providers))
	}
	provider, ok := providers[0].(*searchsvc.PipelineProvider)
	if !ok {
		t.Fatalf("provider type = %T, want *search.PipelineProvider", providers[0])
	}
	secondary, ok := providers[1].(*searchsvc.PipelineProvider)
	if !ok {
		t.Fatalf("second provider type = %T, want *search.PipelineProvider", providers[1])
	}
	if got := secondary.Descriptor().Name; got != "espanol-al-dia" {
		t.Fatalf("second provider name = %q, want espanol-al-dia", got)
	}
	if got := secondary.NormalizerForTesting(); got == nil || got.(*normalize.LiveSearchNormalizer) == nil {
		t.Fatalf("second normalizer type = %T, want *normalize.LiveSearchNormalizer", got)
	}
	if got := provider.FetcherForTesting(); got == nil || got.(*fetch.LiveSearchFetcher) == nil {
		t.Fatalf("fetcher type = %T, want *fetch.LiveSearchFetcher", got)
	}
	if got := provider.ParserForTesting(); got == nil || got.(*parse.LiveSearchParser) == nil {
		t.Fatalf("parser type = %T, want *parse.LiveSearchParser", got)
	}
	if got := provider.NormalizerForTesting(); got == nil || got.(*normalize.LiveSearchNormalizer) == nil {
		t.Fatalf("normalizer type = %T, want *normalize.LiveSearchNormalizer", got)
	}
	wantConfig := config.DefaultRuntimeConfig().Search
	if got := registry.DefaultProviderForTesting(); got != wantConfig.DefaultProviders[0] {
		t.Fatalf("default provider = %q, want %q", got, wantConfig.DefaultProviders[0])
	}
	if got := service.MaxConcurrentForTesting(); got != wantConfig.MaxConcurrent {
		t.Fatalf("max concurrent = %d, want %d", got, wantConfig.MaxConcurrent)
	}
	if _, ok := provider.FetcherForTesting().(*fetch.LiveSearchFetcher); !ok {
		t.Fatalf("fetcher type = %T, want *fetch.LiveSearchFetcher", provider.FetcherForTesting())
	}
}
