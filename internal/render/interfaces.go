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

// EnvelopeRenderer wraps successful outputs, help payloads, and structured fallbacks.
type EnvelopeRenderer interface {
	RenderSuccess(ctx context.Context, env model.Envelope, body []byte) ([]byte, error)
	RenderHelp(ctx context.Context, help model.HelpEnvelope) ([]byte, error)
	RenderFallback(ctx context.Context, fb model.FallbackEnvelope) ([]byte, error)
}
