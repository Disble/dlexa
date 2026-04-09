package espanolaldia

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/render"
)

func TestModuleTargetsEspanolAlDiaSourceAndReturnsBody(t *testing.T) {
	lookup := &lookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "solo", Format: "markdown", Sources: []string{"espanol-al-dia"}}, CacheHit: true, Entries: []model.Entry{{Headword: "solo", Content: "contenido"}}}}
	renderers := &renderersStub{renderer: &rendererStub{payload: []byte("## Español al día\ncontenido")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "markdown"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := lookup.request.Sources; len(got) != 1 || got[0] != "espanol-al-dia" {
		t.Fatalf("lookup sources = %#v, want [\"espanol-al-dia\"]", got)
	}
	if response.Source != "Español al día" || string(response.Body) != "## Español al día\ncontenido" {
		t.Fatalf("response = %#v", response)
	}
}

func TestModuleMapsNotFoundIntoStructuredFallback(t *testing.T) {
	lookup := &lookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "solo", Format: "json", Sources: []string{"espanol-al-dia"}}, Misses: []model.LookupMiss{{Kind: model.LookupMissKindGenericNotFound, Query: "solo"}}}}
	renderers := &renderersStub{renderer: &rendererStub{payload: []byte("{}")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "json", Sources: []string{"espanol-al-dia"}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindNotFound {
		t.Fatalf("fallback = %#v, want not_found fallback", response.Fallback)
	}
	if response.Fallback.NextCommand != "dlexa search solo" {
		t.Fatalf("next command = %q, want dlexa search solo", response.Fallback.NextCommand)
	}
}

func TestModulePreservesJSONLookupSchema(t *testing.T) {
	lookup := &lookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "solo", Format: "json", Sources: []string{"espanol-al-dia"}}, Entries: []model.Entry{{Headword: "solo", Content: "contenido"}}}}
	jsonBody, _ := json.Marshal(lookup.result)
	module := New(lookup, &renderersStub{renderer: &rendererStub{payload: jsonBody}})

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "json", Sources: []string{"espanol-al-dia"}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback != nil {
		t.Fatalf("fallback = %#v, want nil", response.Fallback)
	}
	if !strings.Contains(string(response.Body), `"Entries"`) {
		t.Fatalf("body = %s, want lookup JSON schema", string(response.Body))
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
