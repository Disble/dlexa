// Package modules defines shared contracts for CLI-facing content modules.
package modules

import (
	"context"
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

// Request is the shared application-facing module contract.
type Request struct {
	Query   string
	Format  string
	NoCache bool
	Args    []string
	Sources []string
}

// Response is the shared application-facing module response contract.
type Response struct {
	Title      string
	Source     string
	CacheState string
	Format     string
	Body       []byte
	Fallback   *model.FallbackEnvelope
}

// Module exposes one CLI-capable capability behind a shared contract.
type Module interface {
	Name() string
	Command() string
	Execute(ctx context.Context, req Request) (Response, error)
}

// Registry stores modules by public command name.
type Registry struct {
	modules map[string]Module
}

// NewRegistry builds a module registry.
func NewRegistry(items ...Module) *Registry {
	registry := &Registry{modules: make(map[string]Module, len(items))}
	for _, item := range items {
		if item == nil {
			continue
		}
		registry.modules[strings.TrimSpace(item.Command())] = item
	}
	return registry
}

// Module resolves a module by command name.
func (r *Registry) Module(command string) (Module, bool) {
	if r == nil {
		return nil, false
	}
	module, ok := r.modules[strings.TrimSpace(command)]
	return module, ok
}

// CacheState renders a stable HIT/MISS marker for envelopes.
func CacheState(hit bool) string {
	if hit {
		return "HIT"
	}
	return "MISS"
}

// NotFoundFallback creates a standard Level 2 fallback.
func NotFoundFallback(moduleName, query, nextCommand string) *model.FallbackEnvelope {
	return &model.FallbackEnvelope{
		Kind:        model.FallbackKindNotFound,
		Module:      moduleName,
		Title:       query,
		Query:       query,
		Message:     "No se encontró contenido en este módulo.",
		Suggestion:  "Probá con `dlexa search <consulta>` para descubrir la ruta correcta.",
		NextCommand: nextCommand,
	}
}

// FallbackFromError maps known problem codes into the 4-level fallback ladder.
func FallbackFromError(moduleName, query, format string, err error) *model.FallbackEnvelope {
	problem, ok := model.AsProblem(err)
	if !ok {
		return &model.FallbackEnvelope{
			Kind:       model.FallbackKindParseFailure,
			Module:     moduleName,
			Title:      query,
			Query:      query,
			Format:     format,
			Message:    "La respuesta no pudo interpretarse de forma segura.",
			Detail:     err.Error(),
			Suggestion: "Hace falta intervención humana de mantenimiento.",
		}
	}

	base := &model.FallbackEnvelope{
		Module: moduleName,
		Title:  query,
		Query:  query,
		Format: format,
		Detail: problem.Message,
	}

	switch problem.Code {
	case model.ProblemCodeDPDNotFound, model.ProblemCodeArticleNotFound:
		base.Kind = model.FallbackKindNotFound
		base.Message = "No se encontró contenido en este módulo."
		base.Suggestion = "Probá con `dlexa search <consulta>` para descubrir una ruta válida."
		base.NextCommand = "dlexa search " + query
	case model.ProblemCodeDPDFetchFailed, model.ProblemCodeDPDSearchFetchFailed, model.ProblemCodeArticleFetchFailed, model.ProblemCodeSourceLookupFailed:
		base.Kind = model.FallbackKindUpstreamUnavailable
		base.Message = "La fuente externa no está disponible ahora mismo."
		base.Suggestion = "NO reintentes en loop automático; abortá y reintentá más tarde."
	case model.ProblemCodeDPDExtractFailed, model.ProblemCodeDPDTransformFailed, model.ProblemCodeDPDSearchParseFailed, model.ProblemCodeDPDSearchNormalizeFailed, model.ProblemCodeArticleExtractFailed, model.ProblemCodeArticleTransformFailed:
		base.Kind = model.FallbackKindParseFailure
		base.Message = "La fuente respondió, pero cambió el contrato que esperamos."
		base.Suggestion = "Hace falta intervención humana de mantenimiento."
	default:
		base.Kind = model.FallbackKindParseFailure
		base.Message = "La respuesta no pudo interpretarse de forma segura."
		base.Suggestion = "Hace falta intervención humana de mantenimiento."
	}

	return base
}
