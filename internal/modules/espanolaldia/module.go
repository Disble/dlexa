// Package espanolaldia adapts Español al día lookup behavior to the shared module contract.
package espanolaldia

import (
	"context"

	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/query"
	"github.com/Disble/dlexa/internal/render"
)

const (
	moduleName   = "espanol-al-dia"
	moduleSource = "Español al día"
)

// Module adapts the español-al-día lookup service to the shared module contract.
type Module struct {
	lookup    query.Looker
	renderers render.RendererResolver
}

// New creates an Español al día module.
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
	return modules.ExecuteLookupModule(ctx, req, m.lookup, m.renderers, modules.LookupModuleOptions{
		ModuleName:      moduleName,
		ModuleSource:    moduleSource,
		MissingFallback: modules.DefaultLookupMissingFallback(moduleName),
	})
}
