package app

import (
	"github.com/gentleman-programming/dlexa/internal/cache"
	"github.com/gentleman-programming/dlexa/internal/config"
	"github.com/gentleman-programming/dlexa/internal/doctor"
	"github.com/gentleman-programming/dlexa/internal/fetch"
	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/normalize"
	"github.com/gentleman-programming/dlexa/internal/parse"
	"github.com/gentleman-programming/dlexa/internal/platform"
	"github.com/gentleman-programming/dlexa/internal/query"
	"github.com/gentleman-programming/dlexa/internal/render"
	"github.com/gentleman-programming/dlexa/internal/source"
)

// New is the composition root: concrete adapters are chosen here and nowhere else.
func New(cli platform.CLI) *App {
	runtimeConfig := config.RuntimeConfig{
		DefaultFormat:  "markdown",
		DefaultSources: []string{"demo"},
		CacheEnabled:   true,
	}

	loader := config.NewStaticLoader(runtimeConfig)
	doctorService := doctor.NewNoopDoctor()
	cacheStore := cache.NewMemoryStore()

	demoSource := source.NewPipelineSource(
		model.SourceDescriptor{
			Name:        "demo",
			DisplayName: "Demo Source",
			Kind:        "bootstrap",
			Priority:    1,
			Cacheable:   true,
		},
		fetch.NewStaticFetcher("https://example.invalid/dlexa"),
		parse.NewMarkdownParser(),
		normalize.NewIdentityNormalizer(),
	)

	registry := source.NewStaticRegistry(demoSource)
	lookupService := query.NewService(registry, cacheStore)
	rendererRegistry := render.NewRegistry(
		render.NewMarkdownRenderer(),
		render.NewJSONRenderer(),
	)

	return &App{
		platform:  cli,
		config:    loader,
		doctor:    doctorService,
		lookup:    lookupService,
		renderers: rendererRegistry,
	}
}
