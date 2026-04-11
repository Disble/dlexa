package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Disble/dlexa/internal/cache"
	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/doctor"
	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	moddpd "github.com/Disble/dlexa/internal/modules/dpd"
	moddl "github.com/Disble/dlexa/internal/modules/dudalinguistica"
	modead "github.com/Disble/dlexa/internal/modules/espanolaldia"
	modnoticia "github.com/Disble/dlexa/internal/modules/noticia"
	modsearch "github.com/Disble/dlexa/internal/modules/search"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
	parseengine "github.com/Disble/dlexa/internal/parse/engine"
	"github.com/Disble/dlexa/internal/platform"
	"github.com/Disble/dlexa/internal/query"
	"github.com/Disble/dlexa/internal/render"
	searchsvc "github.com/Disble/dlexa/internal/search"
	"github.com/Disble/dlexa/internal/source"
)

// New is the composition root: concrete adapters are chosen here and nowhere else.
func New(cli platform.CLI) *App {
	runtimeConfig := config.DefaultRuntimeConfig()

	loader := config.NewStaticLoader(runtimeConfig)
	doctorService := doctor.NewNoopDoctor()

	var cacheStore cache.Store
	var searchCacheStore cache.SearchStore
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheStore = cache.NewMemoryStore() // fallback: no persistent cache
		searchCacheStore = cache.NewSearchMemoryStore()
	} else {
		cacheStore = cache.NewFilesystemStore(
			filepath.Join(cacheDir, "dlexa"),
			runtimeConfig.CacheTTL,
		)
		searchCacheStore = cache.NewSearchFilesystemStore(
			filepath.Join(cacheDir, "dlexa", "search"),
			runtimeConfig.CacheTTL,
		)
	}

	dpdSource := source.NewEnginePipelineSource(
		model.SourceDescriptor{
			Name:        "dpd",
			DisplayName: "Diccionario panhispánico de dudas",
			Kind:        "remote-html",
			Priority:    1,
			Cacheable:   true,
		},
		fetch.NewDPDFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent),
		parseengine.NewDPDArticleParser(),
		normalize.NewDPDNormalizer(),
	)

	espanolAlDiaSource := source.NewEnginePipelineSource(
		model.SourceDescriptor{
			Name:        "espanol-al-dia",
			DisplayName: "Español al día",
			Kind:        "remote-html",
			Priority:    2,
			Cacheable:   true,
		},
		fetch.NewEspanolAlDiaFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent),
		parseengine.NewEspanolAlDiaArticleParser(),
		normalize.NewEspanolAlDiaNormalizer(),
	)

	dudaLinguisticaSource := source.NewEnginePipelineSource(
		model.SourceDescriptor{
			Name:        "duda-linguistica",
			DisplayName: "Duda lingüística",
			Kind:        "remote-html",
			Priority:    3,
			Cacheable:   true,
		},
		fetch.NewDudaLinguisticaFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent),
		parseengine.NewDudaLinguisticaArticleParser(),
		normalize.NewDudaLinguisticaNormalizer(),
	)

	noticiaSource := source.NewEnginePipelineSource(
		model.SourceDescriptor{
			Name:        "noticia",
			DisplayName: "Preguntas frecuentes RAE",
			Kind:        "remote-html",
			Priority:    4,
			Cacheable:   true,
		},
		fetch.NewNoticiaFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent),
		parseengine.NewNoticiaArticleParser(),
		normalize.NewNoticiaNormalizer(),
	)

	demoSource := source.NewPipelineSource(
		model.SourceDescriptor{
			Name:        "demo",
			DisplayName: "Demo Source",
			Kind:        "bootstrap",
			Priority:    99,
			Cacheable:   true,
		},
		fetch.NewStaticFetcher("https://example.invalid/dlexa"),
		parse.NewMarkdownParser(),
		normalize.NewIdentityNormalizer(),
	)

	registry := source.NewStaticRegistry(dpdSource, espanolAlDiaSource, dudaLinguisticaSource, noticiaSource, demoSource)
	lookupService := query.NewService(registry, cacheStore)
	governedSearchClient := func(next fetch.Doer) fetch.Doer {
		return fetch.NewGovernedDoer(next, fetch.GovernanceConfig{
			CooldownBase:      runtimeConfig.Search.Governance.CooldownBase,
			CooldownMax:       runtimeConfig.Search.Governance.CooldownMax,
			RespectRetryAfter: runtimeConfig.Search.Governance.RespectRetryAfter,
		})
	}
	searchFetcher := fetch.NewLiveSearchFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent)
	searchFetcher.Client = governedSearchClient(searchFetcher.Client)
	dpdSearchFetcher := fetch.NewDPDSearchFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent)
	dpdSearchFetcher.Client = governedSearchClient(dpdSearchFetcher.Client)
	searchProvider := searchsvc.NewEnginePipelineProvider(
		model.SourceDescriptor{Name: "search", DisplayName: "Búsqueda general RAE", Kind: "remote-html", Priority: 1, Cacheable: true},
		searchFetcher,
		parseengine.NewLiveSearchParser(),
		normalize.NewLiveSearchNormalizer(),
	)
	dpdSearchProvider := searchsvc.NewEnginePipelineProvider(
		model.SourceDescriptor{Name: "dpd", DisplayName: "Diccionario panhispánico de dudas", Kind: "remote-json", Priority: 2, Cacheable: true},
		dpdSearchFetcher,
		parseengine.NewDPDSearchParser(),
		normalize.NewDPDSearchNormalizer(),
	)
	defaultSearchProvider := defaultSearchProvider(runtimeConfig.Search.DefaultProviders)
	searchRegistry := searchsvc.NewStaticRegistry(defaultSearchProvider, searchProvider, dpdSearchProvider)
	searchService := searchsvc.NewService(searchRegistry, searchCacheStore, runtimeConfig.Search.MaxConcurrent, defaultSearchProvider)
	rendererRegistry := render.NewRegistry(
		render.NewMarkdownRenderer(),
		render.NewJSONRenderer(),
	)
	searchRendererRegistry := render.NewSearchRegistry(
		render.NewSearchMarkdownRenderer(),
		render.NewSearchJSONRenderer(),
	)
	moduleRegistry := modules.NewRegistry(
		moddpd.New(lookupService, rendererRegistry),
		modead.New(lookupService, rendererRegistry),
		moddl.New(lookupService, rendererRegistry),
		modnoticia.New(lookupService, rendererRegistry),
		modsearch.New(searchService, searchRendererRegistry),
	)

	return NewWithDependencies(cli, loader, doctorService, moduleRegistry, render.NewEnvelopeRenderer())
}

func defaultSearchProvider(providers []string) string {
	for _, provider := range providers {
		if candidate := strings.TrimSpace(provider); candidate != "" {
			return candidate
		}
	}
	return "search"
}
