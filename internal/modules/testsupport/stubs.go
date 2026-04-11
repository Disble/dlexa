// Package testsupport provides reusable test doubles for module packages.
package testsupport

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/render"
)

// LookupStub records lookup requests and returns configured results.
type LookupStub struct {
	Request model.LookupRequest
	Result  model.LookupResult
	Err     error
}

// Lookup records the request and returns the configured result.
func (s *LookupStub) Lookup(_ context.Context, request model.LookupRequest) (model.LookupResult, error) {
	s.Request = request
	return s.Result, s.Err
}

// RenderersStub resolves a preconfigured renderer for tests.
type RenderersStub struct {
	RendererValue render.Renderer
	Err           error
}

// Renderer returns the configured renderer stub.
func (s *RenderersStub) Renderer(string) (render.Renderer, error) {
	return s.RendererValue, s.Err
}

// RendererStub captures the rendered lookup result and returns a payload.
type RendererStub struct {
	FormatValue string
	Payload     []byte
	Result      model.LookupResult
	Err         error
}

// Format returns the configured output format.
func (s *RendererStub) Format() string {
	if s.FormatValue != "" {
		return s.FormatValue
	}
	return "markdown"
}

// Render records the input result and returns the configured payload.
func (s *RendererStub) Render(_ context.Context, result model.LookupResult) ([]byte, error) {
	s.Result = result
	return s.Payload, s.Err
}
