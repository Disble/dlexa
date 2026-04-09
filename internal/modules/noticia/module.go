// Package noticia adapts FAQ-style noticia lookup behavior to the shared module contract.
package noticia

import (
	"context"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/query"
	"github.com/Disble/dlexa/internal/render"
)

const (
	moduleName   = "noticia"
	moduleSource = "Preguntas frecuentes RAE"
)

// Module adapts FAQ-style noticia lookup behavior to the shared module contract.
type Module struct {
	lookup    query.Looker
	renderers render.RendererResolver
}

// New creates a noticia module.
func New(lookup query.Looker, renderers render.RendererResolver) *Module {
	return &Module{lookup: lookup, renderers: renderers}
}

// LookupForTesting exposes the wired lookup service for app wiring tests.
func (m *Module) LookupForTesting() query.Looker { return m.lookup }

// Name returns the semantic module name.
func (m *Module) Name() string { return moduleName }

// Command returns the public CLI command for this module.
func (m *Module) Command() string { return moduleName }

// Execute resolves the lookup, enforces the FAQ prefix policy, and renders a structured response.
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
	if !isAllowedNoticiaResult(result) {
		return modules.Response{
			Title:      lookupReq.Query,
			Source:     moduleSource,
			CacheState: modules.CacheState(result.CacheHit),
			Format:     lookupReq.Format,
			Fallback: &model.FallbackEnvelope{
				Kind:        model.FallbackKindNotFound,
				Module:      moduleName,
				Title:       lookupReq.Query,
				Query:       lookupReq.Query,
				Format:      lookupReq.Format,
				Message:     "Ese slug no expone una pregunta frecuente normativa compatible con este módulo.",
				Suggestion:  "Probá con `dlexa search <consulta>` para encontrar una FAQ válida o la ruta canónica de Español al día.",
				NextCommand: "dlexa search " + strings.TrimSpace(lookupReq.Query),
			},
		}, nil
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

func isAllowedNoticiaResult(result model.LookupResult) bool {
	if len(result.Entries) == 0 {
		return false
	}
	article := result.Entries[0].Article
	if article == nil {
		return false
	}
	return strings.HasPrefix(normalizeTitle(article.Lemma), faqTitlePrefix)
}

const faqTitlePrefix = "preguntas frecuentes:"

func normalizeTitle(title string) string {
	return strings.ToLower(strings.TrimSpace(title))
}
