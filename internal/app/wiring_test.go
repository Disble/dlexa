package app

import (
	"testing"

	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	moddpd "github.com/Disble/dlexa/internal/modules/dpd"
	moddl "github.com/Disble/dlexa/internal/modules/dudalinguistica"
	modead "github.com/Disble/dlexa/internal/modules/espanolaldia"
	modnoticia "github.com/Disble/dlexa/internal/modules/noticia"
	modsearch "github.com/Disble/dlexa/internal/modules/search"
	"github.com/Disble/dlexa/internal/normalize"
	parseengine "github.com/Disble/dlexa/internal/parse/engine"
	"github.com/Disble/dlexa/internal/query"
	searchsvc "github.com/Disble/dlexa/internal/search"
	"github.com/Disble/dlexa/internal/source"
)

func TestNewWiresDPDModuleToEngineArticleParser(t *testing.T) {
	application := New(&fakeCLI{})
	module, ok := application.registry.Module("dpd")
	if !ok {
		t.Fatal("dpd module not registered")
	}
	dpdModule, ok := module.(*moddpd.Module)
	if !ok {
		t.Fatalf("module type = %T, want *dpd.Module", module)
	}
	lookupForTesting := dpdModule.LookupForTesting()
	lookup, ok := lookupForTesting.(*query.LookupService)
	if !ok {
		t.Fatalf(errLookupType, lookupForTesting)
	}
	registryForTesting := lookup.RegistryForTesting()
	registry, ok := registryForTesting.(*source.StaticRegistry)
	if !ok {
		t.Fatalf(errRegistryType, registryForTesting)
	}
	resolved, err := registry.SourcesFor(model.LookupRequest{Query: "solo", Sources: []string{"dpd"}})
	if err != nil {
		t.Fatalf(errSourcesFor, err)
	}
	if len(resolved) != 1 {
		t.Fatalf(errResolvedSourcesLen, len(resolved))
	}
	pipeline, ok := resolved[0].(*source.PipelineSource)
	if !ok {
		t.Fatalf(errPipelineSourceType, resolved[0])
	}
	if _, ok := pipeline.ArticleEngineParserForTesting().(*parseengine.DPDArticleParser); !ok {
		t.Fatalf("article engine parser type = %T, want *engine.DPDArticleParser", pipeline.ArticleEngineParserForTesting())
	}
}

func TestNewWiresSearchModuleToLiveSearchAdapters(t *testing.T) {
	application := New(&fakeCLI{})
	service, registry := requireSearchServiceAndRegistry(t, application)
	providers := registry.Providers()
	if len(providers) != 2 {
		t.Fatalf("providers len = %d, want 2", len(providers))
	}
	provider := requireSearchPipelineProvider(t, providers[0], "provider")
	secondary := requireSearchPipelineProvider(t, providers[1], "second provider")
	assertLiveSearchPipelineProvider(t, provider)
	assertDPDSearchPipelineProvider(t, secondary)
	assertSearchModuleRuntimeConfig(t, service, registry)
}

func TestNewWiresEspanolAlDiaModuleToEngineArticleParser(t *testing.T) {
	application := New(&fakeCLI{})
	module, ok := application.registry.Module(commandEspanolAlDia)
	if !ok {
		t.Fatalf("%s module not registered", commandEspanolAlDia)
	}
	eadModule, ok := module.(*modead.Module)
	if !ok {
		t.Fatalf("module type = %T, want *espanolaldia.Module", module)
	}
	lookupForTesting := eadModule.LookupForTesting()
	lookup, ok := lookupForTesting.(*query.LookupService)
	if !ok {
		t.Fatalf(errLookupType, lookupForTesting)
	}
	registryForTesting := lookup.RegistryForTesting()
	registry, ok := registryForTesting.(*source.StaticRegistry)
	if !ok {
		t.Fatalf(errRegistryType, registryForTesting)
	}
	resolved, err := registry.SourcesFor(model.LookupRequest{Query: "solo", Sources: []string{commandEspanolAlDia}})
	if err != nil {
		t.Fatalf(errSourcesFor, err)
	}
	if len(resolved) != 1 {
		t.Fatalf(errResolvedSourcesLen, len(resolved))
	}
	pipeline, ok := resolved[0].(*source.PipelineSource)
	if !ok {
		t.Fatalf(errPipelineSourceType, resolved[0])
	}
	if _, ok := pipeline.ArticleEngineParserForTesting().(*parseengine.EspanolAlDiaArticleParser); !ok {
		t.Fatalf("article engine parser type = %T, want *engine.EspanolAlDiaArticleParser", pipeline.ArticleEngineParserForTesting())
	}
	if got := pipeline.Descriptor().Name; got != commandEspanolAlDia {
		t.Fatalf("descriptor name = %q, want %s", got, commandEspanolAlDia)
	}
	if got := pipeline.Descriptor().DisplayName; got != "Español al día" {
		t.Fatalf("descriptor display name = %q, want Español al día", got)
	}
	if got := resolved[0].Descriptor().Priority; got != 2 {
		t.Fatalf("priority = %d, want 2", got)
	}
	if fetch.NewEspanolAlDiaFetcher(config.DefaultDPDBaseURL, config.DefaultDPDTimeout, config.DefaultDPDUserAgent) == nil {
		t.Fatal("expected concrete espanol-al-dia fetcher constructor to return instance")
	}
}

func TestNewWiresDudaLinguisticaModuleToEngineArticleParser(t *testing.T) {
	application := New(&fakeCLI{})
	module, ok := application.registry.Module(commandDudaLinguistica)
	if !ok {
		t.Fatalf("%s module not registered", commandDudaLinguistica)
	}
	dlModule, ok := module.(*moddl.Module)
	if !ok {
		t.Fatalf("module type = %T, want *dudalinguistica.Module", module)
	}
	lookupForTesting := dlModule.LookupForTesting()
	lookup, ok := lookupForTesting.(*query.LookupService)
	if !ok {
		t.Fatalf(errLookupType, lookupForTesting)
	}
	registryForTesting := lookup.RegistryForTesting()
	registry, ok := registryForTesting.(*source.StaticRegistry)
	if !ok {
		t.Fatalf(errRegistryType, registryForTesting)
	}
	resolved, err := registry.SourcesFor(model.LookupRequest{Query: "tilde", Sources: []string{commandDudaLinguistica}})
	if err != nil {
		t.Fatalf(errSourcesFor, err)
	}
	if len(resolved) != 1 {
		t.Fatalf(errResolvedSourcesLen, len(resolved))
	}
	pipeline, ok := resolved[0].(*source.PipelineSource)
	if !ok {
		t.Fatalf(errPipelineSourceType, resolved[0])
	}
	if _, ok := pipeline.ArticleEngineParserForTesting().(*parseengine.DudaLinguisticaArticleParser); !ok {
		t.Fatalf("article engine parser type = %T, want *engine.DudaLinguisticaArticleParser", pipeline.ArticleEngineParserForTesting())
	}
	if got := pipeline.Descriptor().Name; got != commandDudaLinguistica {
		t.Fatalf("descriptor name = %q, want %s", got, commandDudaLinguistica)
	}
	if got := pipeline.Descriptor().DisplayName; got != "Duda lingüística" {
		t.Fatalf("descriptor display name = %q, want Duda lingüística", got)
	}
	if got := resolved[0].Descriptor().Priority; got != 3 {
		t.Fatalf("priority = %d, want 3", got)
	}
	if fetch.NewDudaLinguisticaFetcher(config.DefaultDPDBaseURL, config.DefaultDPDTimeout, config.DefaultDPDUserAgent) == nil {
		t.Fatal("expected concrete duda-linguistica fetcher constructor to return instance")
	}
}

func TestNewWiresNoticiaModuleToEngineArticleParser(t *testing.T) {
	application := New(&fakeCLI{})
	module, ok := application.registry.Module("noticia")
	if !ok {
		t.Fatal("noticia module not registered")
	}
	noticiaModule, ok := module.(*modnoticia.Module)
	if !ok {
		t.Fatalf("module type = %T, want *noticia.Module", module)
	}
	lookupForTesting := noticiaModule.LookupForTesting()
	lookup, ok := lookupForTesting.(*query.LookupService)
	if !ok {
		t.Fatalf(errLookupType, lookupForTesting)
	}
	registryForTesting := lookup.RegistryForTesting()
	registry, ok := registryForTesting.(*source.StaticRegistry)
	if !ok {
		t.Fatalf(errRegistryType, registryForTesting)
	}
	resolved, err := registry.SourcesFor(model.LookupRequest{Query: "preguntas-frecuentes-tilde-en-las-mayusculas", Sources: []string{"noticia"}})
	if err != nil {
		t.Fatalf(errSourcesFor, err)
	}
	if len(resolved) != 1 {
		t.Fatalf(errResolvedSourcesLen, len(resolved))
	}
	pipeline, ok := resolved[0].(*source.PipelineSource)
	if !ok {
		t.Fatalf(errPipelineSourceType, resolved[0])
	}
	if _, ok := pipeline.ArticleEngineParserForTesting().(*parseengine.NoticiaArticleParser); !ok {
		t.Fatalf("article engine parser type = %T, want *engine.NoticiaArticleParser", pipeline.ArticleEngineParserForTesting())
	}
	if got := pipeline.Descriptor().Name; got != "noticia" {
		t.Fatalf("descriptor name = %q, want noticia", got)
	}
	if got := pipeline.Descriptor().DisplayName; got != "Preguntas frecuentes RAE" {
		t.Fatalf("descriptor display name = %q, want Preguntas frecuentes RAE", got)
	}
	if got := resolved[0].Descriptor().Priority; got != 4 {
		t.Fatalf("priority = %d, want 4", got)
	}
	if fetch.NewNoticiaFetcher(config.DefaultDPDBaseURL, config.DefaultDPDTimeout, config.DefaultDPDUserAgent) == nil {
		t.Fatal("expected concrete noticia fetcher constructor to return instance")
	}
}

func requireSearchServiceAndRegistry(t *testing.T, application *App) (*searchsvc.Service, *searchsvc.StaticRegistry) {
	t.Helper()

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
	return service, registry
}

func requireSearchPipelineProvider(t *testing.T, provider any, label string) *searchsvc.PipelineProvider {
	t.Helper()

	pipelineProvider, ok := provider.(*searchsvc.PipelineProvider)
	if !ok {
		t.Fatalf("%s type = %T, want *search.PipelineProvider", label, provider)
	}
	return pipelineProvider
}

func assertLiveSearchPipelineProvider(t *testing.T, provider *searchsvc.PipelineProvider) {
	t.Helper()

	if got := provider.FetcherForTesting(); got == nil {
		t.Fatal("fetcher = nil, want *fetch.LiveSearchFetcher")
	} else if _, ok := got.(*fetch.LiveSearchFetcher); !ok {
		t.Fatalf("fetcher type = %T, want *fetch.LiveSearchFetcher", got)
	}
	if _, ok := provider.EngineSearchParserForTesting().(*parseengine.LiveSearchParser); !ok {
		t.Fatalf("engine parser type = %T, want *engine.LiveSearchParser", provider.EngineSearchParserForTesting())
	}
	if got := provider.NormalizerForTesting(); got == nil {
		t.Fatal("normalizer = nil, want *normalize.LiveSearchNormalizer")
	} else if _, ok := got.(*normalize.LiveSearchNormalizer); !ok {
		t.Fatalf("normalizer type = %T, want *normalize.LiveSearchNormalizer", got)
	}
}

func assertDPDSearchPipelineProvider(t *testing.T, provider *searchsvc.PipelineProvider) {
	t.Helper()

	if got := provider.Descriptor().Name; got != "dpd" {
		t.Fatalf("second provider name = %q, want dpd", got)
	}
	fetcher, ok := provider.FetcherForTesting().(*fetch.DPDSearchFetcher)
	if !ok || fetcher == nil {
		t.Fatalf("second fetcher type = %T, want *fetch.DPDSearchFetcher", provider.FetcherForTesting())
	}
	if _, ok := fetcher.Client.(*fetch.GovernedDoer); !ok {
		t.Fatalf("second fetcher client = %T, want *fetch.GovernedDoer", fetcher.Client)
	}
	if _, ok := provider.EngineSearchParserForTesting().(*parseengine.DPDSearchParser); !ok {
		t.Fatalf("second engine parser type = %T, want *engine.DPDSearchParser", provider.EngineSearchParserForTesting())
	}
	if got := provider.NormalizerForTesting(); got == nil {
		t.Fatal("second normalizer = nil, want *normalize.DPDSearchNormalizer")
	} else if _, ok := got.(*normalize.DPDSearchNormalizer); !ok {
		t.Fatalf("second normalizer type = %T, want *normalize.DPDSearchNormalizer", got)
	}
}

func assertSearchModuleRuntimeConfig(t *testing.T, service *searchsvc.Service, registry *searchsvc.StaticRegistry) {
	t.Helper()

	wantConfig := config.DefaultRuntimeConfig().Search
	if got := registry.DefaultProviderForTesting(); got != wantConfig.DefaultProviders[0] {
		t.Fatalf("default provider = %q, want %q", got, wantConfig.DefaultProviders[0])
	}
	if got := service.MaxConcurrentForTesting(); got != wantConfig.MaxConcurrent {
		t.Fatalf("max concurrent = %d, want %d", got, wantConfig.MaxConcurrent)
	}
}
