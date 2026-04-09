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
	moduleName          = "search"
	generalSearchSource = "búsqueda general RAE"
	dpdSearchSource     = "Diccionario panhispánico de dudas"
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

// SearcherForTesting exposes the wired searcher for black-box wiring tests.
func (m *Module) SearcherForTesting() searchsvc.Searcher { return m.searcher }

// Execute curates the upstream search result and maps semantic URLs into actionable commands.
func (m *Module) Execute(ctx context.Context, req modules.Request) (modules.Response, error) {
	searchReq := model.SearchRequest{Query: strings.TrimSpace(req.Query), Format: strings.TrimSpace(req.Format), Sources: append([]string(nil), req.Sources...), NoCache: req.NoCache}
	responseSource := sourceLabelForProviders(searchReq.Sources)
	result, err := m.searcher.Search(ctx, searchReq)
	if err != nil {
		return modules.Response{Title: searchReq.Query, Source: responseSource, CacheState: modules.CacheState(false), Format: searchReq.Format, Fallback: modules.FallbackFromError(moduleName, searchReq.Query, searchReq.Format, err)}, nil
	}
	result.Candidates = curateCandidates(searchReq.Query, result.Candidates)
	if len(result.Candidates) == 0 {
		result.Outcome = model.SearchOutcomeNoResults
	} else {
		result.Outcome = model.SearchOutcomeResults
	}
	renderer, err := m.renderers.Renderer(searchReq.Format)
	if err != nil {
		return modules.Response{}, err
	}
	body, err := renderer.Render(ctx, result)
	if err != nil {
		return modules.Response{}, err
	}
	return modules.Response{Title: searchReq.Query, Source: sourceLabelForProviders(result.Request.Sources), CacheState: modules.CacheState(result.CacheHit), Format: searchReq.Format, Body: body}, nil
}

func sourceLabelForProviders(sources []string) string {
	trimmed := make([]string, 0, len(sources))
	for _, source := range sources {
		if candidate := strings.TrimSpace(source); candidate != "" {
			trimmed = append(trimmed, candidate)
		}
	}
	if len(trimmed) == 0 {
		return generalSearchSource
	}
	if len(trimmed) > 1 {
		return generalSearchSource
	}
	return sourceDisplayLabel(trimmed[0])
}

func sourceDisplayLabel(source string) string {
	switch strings.TrimSpace(source) {
	case "dpd":
		return dpdSearchSource
	case "search":
		return generalSearchSource
	default:
		return strings.TrimSpace(source)
	}
}
