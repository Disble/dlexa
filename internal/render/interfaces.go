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

// RendererResolver resolves a Renderer by output format name.
type RendererResolver interface {
	Renderer(format string) (Renderer, error)
}
