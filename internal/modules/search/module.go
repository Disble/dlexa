package search

import (
	"context"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/render"
	searchsvc "github.com/Disble/dlexa/internal/search"
)

const (
	moduleName   = "search"
	moduleSource = "búsqueda general RAE"
)

// Module adapts semantic search to the shared module contract.
type Module struct {
	searcher  searchsvc.Searcher
	renderers render.SearchRendererResolver
}

// New creates a Search module.
func New(searcher searchsvc.Searcher, renderers render.SearchRendererResolver) *Module {
	return &Module{searcher: searcher, renderers: renderers}
}

// Name returns the semantic module name.
func (m *Module) Name() string { return moduleName }

// Command returns the public CLI command.
func (m *Module) Command() string { return moduleName }

// Execute curates the upstream search result and maps semantic URLs into actionable commands.
func (m *Module) Execute(ctx context.Context, req modules.Request) (modules.Response, error) {
	searchReq := model.SearchRequest{Query: strings.TrimSpace(req.Query), Format: strings.TrimSpace(req.Format), NoCache: req.NoCache}
	result, err := m.searcher.Search(ctx, searchReq)
	if err != nil {
		return modules.Response{Title: searchReq.Query, Source: moduleSource, CacheState: modules.CacheState(false), Format: searchReq.Format, Fallback: modules.FallbackFromError(moduleName, searchReq.Query, searchReq.Format, err)}, nil
	}
	result.Candidates = curateCandidates(searchReq.Query, result.Candidates)
	renderer, err := m.renderers.Renderer(searchReq.Format)
	if err != nil {
		return modules.Response{}, err
	}
	body, err := renderer.Render(ctx, result)
	if err != nil {
		return modules.Response{}, err
	}
	response := modules.Response{Title: searchReq.Query, Source: moduleSource, CacheState: modules.CacheState(result.CacheHit), Format: searchReq.Format, Body: body}
	if len(result.Candidates) == 0 {
		response.Fallback = modules.NotFoundFallback(moduleName, searchReq.Query, "dlexa search "+searchReq.Query)
	}
	return response, nil
}
