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
	return modules.ExecuteLookupModule(ctx, req, m.lookup, m.renderers, modules.LookupModuleOptions{
		ModuleName:      moduleName,
		ModuleSource:    moduleSource,
		MissingFallback: modules.DefaultLookupMissingFallback(moduleName),
		ResultFallback:  noticiaResultFallback,
	})
}

func noticiaResultFallback(req model.LookupRequest, result model.LookupResult) *model.FallbackEnvelope {
	if isAllowedNoticiaResult(result) {
		return nil
	}
	return &model.FallbackEnvelope{
		Kind:        model.FallbackKindNotFound,
		Module:      moduleName,
		Title:       req.Query,
		Query:       req.Query,
		Format:      req.Format,
		Message:     "Ese slug no expone una pregunta frecuente normativa compatible con este módulo.",
		Suggestion:  "Probá con `dlexa search <consulta>` para encontrar una FAQ válida o la ruta canónica de Español al día.",
		NextCommand: "dlexa search " + strings.TrimSpace(req.Query),
	}
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
