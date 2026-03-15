package render

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
)

// Renderer converts a LookupResult into a formatted output.
type Renderer interface {
	Format() string
	Render(ctx context.Context, result model.LookupResult) ([]byte, error)
}

// Registry resolves a Renderer by output format name.
type Registry interface {
	Renderer(format string) (Renderer, error)
}
