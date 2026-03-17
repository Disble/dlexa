package app

import (
	"os"
	"path/filepath"

	"github.com/Disble/dlexa/internal/cache"
	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/doctor"
	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
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
	searchService := searchsvc.NewService(
		model.SourceDescriptor{Name: "dpd", DisplayName: "Diccionario panhispánico de dudas", Kind: "remote-json", Priority: 1, Cacheable: true},
		fetch.NewDPDSearchFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent),
		parse.NewDPDSearchParser(),
		normalize.NewDPDSearchNormalizer(),
		searchCacheStore,
	)
	rendererRegistry := render.NewRegistry(
		render.NewMarkdownRenderer(),
		render.NewJSONRenderer(),
	)
	searchRendererRegistry := render.NewSearchRegistry(
		render.NewSearchMarkdownRenderer(),
		render.NewSearchJSONRenderer(),
	)

	return &App{
		platform:        cli,
		config:          loader,
		doctor:          doctorService,
		lookup:          lookupService,
		search:          searchService,
		renderers:       rendererRegistry,
		searchRenderers: searchRendererRegistry,
	}
}
