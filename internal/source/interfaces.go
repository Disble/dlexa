// Package source defines the abstraction for dictionary data sources.
package source

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
)

// Source represents a dictionary lookup backend that can resolve queries into entries.
type Source interface {
	Descriptor() model.SourceDescriptor
	Lookup(ctx context.Context, request model.LookupRequest) (model.SourceResult, error)
}

// Registry provides access to available sources for a given lookup request.
type Registry interface {
	SourcesFor(request model.LookupRequest) ([]Source, error)
}
