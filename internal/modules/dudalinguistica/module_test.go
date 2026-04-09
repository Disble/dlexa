package dudalinguistica

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/render"
)

func TestModuleTargetsDudaLinguisticaSourceAndReturnsBody(t *testing.T) {
	lookup := &lookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "tilde", Format: "markdown", Sources: []string{"duda-linguistica"}}, CacheHit: true, Entries: []model.Entry{{Headword: "tilde", Content: "contenido"}}}}
	renderers := &renderersStub{renderer: &rendererStub{payload: []byte("## Duda lingüística\ncontenido")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "tilde", Format: "markdown"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := lookup.request.Sources; len(got) != 1 || got[0] != "duda-linguistica" {
		t.Fatalf("lookup sources = %#v, want [\"duda-linguistica\"]", got)
	}
	if response.Source != "Duda lingüística" || string(response.Body) != "## Duda lingüística\ncontenido" {
		t.Fatalf("response = %#v", response)
	}
}

func TestModuleMapsNotFoundIntoStructuredFallback(t *testing.T) {
	lookup := &lookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "tilde", Format: "markdown", Sources: []string{"duda-linguistica"}}, Misses: []model.LookupMiss{{Kind: model.LookupMissKindGenericNotFound, Query: "tilde"}}}}
	renderers := &renderersStub{renderer: &rendererStub{payload: []byte("unused")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "tilde", Format: "markdown", Sources: []string{"duda-linguistica"}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindNotFound {
		t.Fatalf("fallback = %#v, want not_found fallback", response.Fallback)
	}
	if !strings.Contains(response.Fallback.NextCommand, "dlexa search tilde") {
		t.Fatalf("next command = %q, want dlexa search tilde", response.Fallback.NextCommand)
	}
}

type lookupStub struct {
	request model.LookupRequest
	result  model.LookupResult
	err     error
}

func (s *lookupStub) Lookup(_ context.Context, request model.LookupRequest) (model.LookupResult, error) {
	s.request = request
	return s.result, s.err
}

type renderersStub struct{ renderer render.Renderer }

func (s *renderersStub) Renderer(string) (render.Renderer, error) { return s.renderer, nil }

type rendererStub struct {
	payload []byte
	result  model.LookupResult
}

func (s *rendererStub) Format() string { return "markdown" }
func (s *rendererStub) Render(_ context.Context, result model.LookupResult) ([]byte, error) {
	s.result = result
	return s.payload, nil
}
