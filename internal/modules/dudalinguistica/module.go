// Package dudalinguistica adapts Duda lingüística lookup behavior to the shared module contract.
package dudalinguistica

import (
	"context"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/query"
	"github.com/Disble/dlexa/internal/render"
)

const (
	moduleName   = "duda-linguistica"
	moduleSource = "Duda lingüística"
)

// Module adapts the duda-linguistica lookup service to the shared module contract.
type Module struct {
	lookup    query.Looker
	renderers render.RendererResolver
}

// New creates a Duda lingüística module.
func New(lookup query.Looker, renderers render.RendererResolver) *Module {
	return &Module{lookup: lookup, renderers: renderers}
}

// LookupForTesting exposes the wired lookup service for app wiring tests.
func (m *Module) LookupForTesting() query.Looker { return m.lookup }

// Name returns the semantic module name.
func (m *Module) Name() string { return moduleName }

// Command returns the public CLI command for this module.
func (m *Module) Command() string { return moduleName }

// Execute resolves the lookup and renders a structured lookup response.
func (m *Module) Execute(ctx context.Context, req modules.Request) (modules.Response, error) {
	lookupReq := model.LookupRequest{
		Query:   strings.TrimSpace(req.Query),
		Format:  strings.TrimSpace(req.Format),
		Sources: append([]string(nil), req.Sources...),
		NoCache: req.NoCache,
	}
	if len(lookupReq.Sources) == 0 {
		lookupReq.Sources = []string{moduleName}
	}
	result, err := m.lookup.Lookup(ctx, lookupReq)
	if err != nil {
		return modules.Response{Title: lookupReq.Query, Source: moduleSource, CacheState: modules.CacheState(false), Format: lookupReq.Format, Fallback: modules.FallbackFromError(moduleName, lookupReq.Query, lookupReq.Format, err)}, nil
	}
	renderer, err := m.renderers.Renderer(lookupReq.Format)
	if err != nil {
		return modules.Response{}, err
	}
	body, err := renderer.Render(ctx, result)
	if err != nil {
		return modules.Response{}, err
	}
	response := modules.Response{Title: lookupReq.Query, Source: moduleSource, CacheState: modules.CacheState(result.CacheHit), Format: lookupReq.Format, Body: body}
	if len(result.Entries) == 0 && len(result.Misses) > 0 {
		response.Fallback = modules.NotFoundFallback(moduleName, lookupReq.Query, "dlexa search "+strings.TrimSpace(lookupReq.Query))
	}
	return response, nil
}
