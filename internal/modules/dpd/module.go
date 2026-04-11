// Package dpd adapts DPD lookup behavior to the shared module contract.
package dpd

import (
	"context"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/query"
	"github.com/Disble/dlexa/internal/render"
)

const (
	moduleName   = "dpd"
	moduleSource = "Diccionario panhispánico de dudas"
)

// Module adapts the lookup service to the shared module contract.
type Module struct {
	lookup    query.Looker
	renderers render.RendererResolver
}

// New creates a DPD module.
func New(lookup query.Looker, renderers render.RendererResolver) *Module {
	return &Module{lookup: lookup, renderers: renderers}
}

// LookupForTesting exposes the wired lookup service for app wiring tests.
func (m *Module) LookupForTesting() query.Looker { return m.lookup }

// Name returns the semantic module name.
func (m *Module) Name() string { return moduleName }

// Command returns the public CLI command for this module.
func (m *Module) Command() string { return moduleName }

// Execute resolves the lookup, preserving JSON compatibility while offloading final envelope rendering.
func (m *Module) Execute(ctx context.Context, req modules.Request) (modules.Response, error) {
	return modules.ExecuteLookupModule(ctx, req, m.lookup, m.renderers, modules.LookupModuleOptions{
		ModuleName:   moduleName,
		ModuleSource: moduleSource,
		MissingFallback: func(lookupReq model.LookupRequest, result model.LookupResult) *model.FallbackEnvelope {
			return fallbackFromMiss(lookupReq, result.Misses[0])
		},
	})
}

func fallbackFromMiss(req model.LookupRequest, miss model.LookupMiss) *model.FallbackEnvelope {
	nextCommand := "dlexa search " + strings.TrimSpace(req.Query)
	if miss.NextAction != nil && strings.TrimSpace(miss.NextAction.Command) != "" {
		nextCommand = strings.TrimSpace(miss.NextAction.Command)
	}
	return modules.NotFoundFallback(moduleName, miss.Query, nextCommand)
}
