package render

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
)

// SearchRenderer converts a SearchResult into a formatted output.
type SearchRenderer interface {
	Format() string
	Render(ctx context.Context, result model.SearchResult) ([]byte, error)
}

// SearchRendererResolver resolves a SearchRenderer by output format name.
type SearchRendererResolver interface {
	Renderer(format string) (SearchRenderer, error)
}
