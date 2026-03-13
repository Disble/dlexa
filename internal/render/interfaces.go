package render

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type Renderer interface {
	Format() string
	Render(ctx context.Context, result model.LookupResult) ([]byte, error)
}

type Registry interface {
	Renderer(format string) (Renderer, error)
}
