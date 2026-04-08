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
	modsearch "github.com/Disble/dlexa/internal/modules/search"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
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

	dpdSource := source.NewPipelineSource(
		model.SourceDescriptor{
			Name:        "dpd",
			DisplayName: "Diccionario panhispánico de dudas",
			Kind:        "remote-html",
			Priority:    1,
			Cacheable:   true,
		},
		fetch.NewDPDFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent),
		parse.NewDPDArticleParser(),
		normalize.NewDPDNormalizer(),
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

	registry := source.NewStaticRegistry(dpdSource, demoSource)
	lookupService := query.NewService(registry, cacheStore)
	searchFetcher := fetch.NewLiveSearchFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent)
	searchFetcher.Client = fetch.NewGovernedDoer(searchFetcher.Client, fetch.GovernanceConfig{
		CooldownBase:      runtimeConfig.Search.Governance.CooldownBase,
		CooldownMax:       runtimeConfig.Search.Governance.CooldownMax,
		RespectRetryAfter: runtimeConfig.Search.Governance.RespectRetryAfter,
	})
	searchProvider := searchsvc.NewPipelineProvider(
		model.SourceDescriptor{Name: "search", DisplayName: "Búsqueda general RAE", Kind: "remote-html", Priority: 1, Cacheable: true},
		searchFetcher,
		parse.NewLiveSearchParser(),
		normalize.NewLiveSearchNormalizer(),
	)
	espanolAlDiaProvider := searchsvc.NewPipelineProvider(
		model.SourceDescriptor{Name: "espanol-al-dia", DisplayName: "Español al día", Kind: "remote-html", Priority: 2, Cacheable: true},
		searchFetcher,
		parse.NewLiveSearchParser(),
		normalize.NewLiveSearchNormalizer(),
	)
	defaultSearchProvider := defaultSearchProvider(runtimeConfig.Search.DefaultProviders)
	searchRegistry := searchsvc.NewStaticRegistry(defaultSearchProvider, searchProvider, espanolAlDiaProvider)
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
