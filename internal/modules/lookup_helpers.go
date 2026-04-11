package modules

import (
	"context"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/query"
	"github.com/Disble/dlexa/internal/render"
)

// LookupModuleOptions configures shared lookup-module execution behavior.
type LookupModuleOptions struct {
	ModuleName      string
	ModuleSource    string
	MissingFallback func(model.LookupRequest, model.LookupResult) *model.FallbackEnvelope
	ResultFallback  func(model.LookupRequest, model.LookupResult) *model.FallbackEnvelope
}

// BuildLookupRequest normalizes a module request into a lookup request.
func BuildLookupRequest(req Request, defaultSource string) model.LookupRequest {
	lookupReq := model.LookupRequest{
		Query:   strings.TrimSpace(req.Query),
		Format:  strings.TrimSpace(req.Format),
		Sources: append([]string(nil), req.Sources...),
		NoCache: req.NoCache,
	}
	if len(lookupReq.Sources) == 0 && strings.TrimSpace(defaultSource) != "" {
		lookupReq.Sources = []string{strings.TrimSpace(defaultSource)}
	}
	return lookupReq
}

// DefaultLookupMissingFallback returns the standard not-found fallback builder.
func DefaultLookupMissingFallback(moduleName string) func(model.LookupRequest, model.LookupResult) *model.FallbackEnvelope {
	return func(req model.LookupRequest, _ model.LookupResult) *model.FallbackEnvelope {
		return NotFoundFallback(moduleName, req.Query, "dlexa search "+strings.TrimSpace(req.Query))
	}
}

// ExecuteLookupModule runs the common lookup-module flow and renders the response.
func ExecuteLookupModule(ctx context.Context, req Request, lookup query.Looker, renderers render.RendererResolver, options LookupModuleOptions) (Response, error) {
	lookupReq := BuildLookupRequest(req, options.ModuleName)
	result, err := lookup.Lookup(ctx, lookupReq)
	if err != nil {
		return Response{Title: lookupReq.Query, Source: options.ModuleSource, CacheState: CacheState(false), Format: lookupReq.Format, Fallback: FallbackFromError(options.ModuleName, lookupReq.Query, lookupReq.Format, err)}, nil
	}
	if options.ResultFallback != nil {
		if fallback := options.ResultFallback(lookupReq, result); fallback != nil {
			return Response{Title: lookupReq.Query, Source: options.ModuleSource, CacheState: CacheState(result.CacheHit), Format: lookupReq.Format, Fallback: fallback}, nil
		}
	}
	renderer, err := renderers.Renderer(lookupReq.Format)
	if err != nil {
		return Response{}, err
	}
	body, err := renderer.Render(ctx, result)
	if err != nil {
		return Response{}, err
	}
	response := Response{Title: lookupReq.Query, Source: options.ModuleSource, CacheState: CacheState(result.CacheHit), Format: lookupReq.Format, Body: body}
	if len(result.Entries) == 0 && len(result.Misses) > 0 {
		fallbackFn := options.MissingFallback
		if fallbackFn == nil {
			fallbackFn = DefaultLookupMissingFallback(options.ModuleName)
		}
		response.Fallback = fallbackFn(lookupReq, result)
	}
	return response, nil
}
