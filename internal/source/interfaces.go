package source

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type Source interface {
	Descriptor() model.SourceDescriptor
	Lookup(ctx context.Context, request model.LookupRequest) (model.SourceResult, error)
}

type Registry interface {
	SourcesFor(request model.LookupRequest) ([]Source, error)
}
