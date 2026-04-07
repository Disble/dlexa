package app

import (
	"testing"

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
	if got := service.FetcherForTesting(); got == nil || got.(*fetch.LiveSearchFetcher) == nil {
		t.Fatalf("fetcher type = %T, want *fetch.LiveSearchFetcher", got)
	}
	if got := service.ParserForTesting(); got == nil || got.(*parse.LiveSearchParser) == nil {
		t.Fatalf("parser type = %T, want *parse.LiveSearchParser", got)
	}
	if got := service.NormalizerForTesting(); got == nil || got.(*normalize.LiveSearchNormalizer) == nil {
		t.Fatalf("normalizer type = %T, want *normalize.LiveSearchNormalizer", got)
	}
}
