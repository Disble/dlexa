package dpd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/render"
)

func TestModuleTranslatesRequestsAndReturnsStructuredFallbacks(t *testing.T) {
	lookup := &dpdLookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "solo", Format: "markdown"}, CacheHit: true, Entries: []model.Entry{{Headword: "solo", Content: "contenido"}}}}
	renderers := &dpdRenderersStub{renderer: &dpdRendererStub{payload: []byte("## DPD\ncontenido")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "markdown", NoCache: true, Sources: []string{"dpd"}})
	if err != nil {
		t.Fatalf("Execute() success error = %v", err)
	}
	if lookup.request.Query != "solo" || lookup.request.Format != "markdown" || !lookup.request.NoCache {
		t.Fatalf("lookup request = %#v", lookup.request)
	}
	if response.Title != "solo" || response.CacheState != modules.CacheState(true) {
		t.Fatalf("response = %#v", response)
	}
	if string(response.Body) != "## DPD\ncontenido" || response.Fallback != nil {
		t.Fatalf("response = %#v", response)
	}

	lookup.result = model.LookupResult{Request: model.LookupRequest{Query: "inexistente", Format: "markdown"}, Misses: []model.LookupMiss{{Kind: model.LookupMissKindGenericNotFound, Query: "inexistente", NextAction: &model.LookupNextAction{Kind: model.LookupNextActionKindSearch, Query: "inexistente", Command: "dlexa search inexistente"}}}}
	renderers.renderer = &dpdRendererStub{payload: []byte("json-compatible-body")}
	response, err = module.Execute(context.Background(), modules.Request{Query: "inexistente", Format: "json", Sources: []string{"dpd"}})
	if err != nil {
		t.Fatalf("Execute() miss error = %v", err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindNotFound || response.Fallback.NextCommand != "dlexa search inexistente" {
		t.Fatalf("fallback = %#v", response.Fallback)
	}
	if string(response.Body) != "json-compatible-body" {
		t.Fatalf("body = %q, want rendered json-compatible payload", string(response.Body))
	}
}

type dpdLookupStub struct {
	request model.LookupRequest
	result  model.LookupResult
	err     error
}

func (s *dpdLookupStub) Lookup(_ context.Context, request model.LookupRequest) (model.LookupResult, error) {
	s.request = request
	return s.result, s.err
}

type dpdRenderersStub struct{ renderer render.Renderer }

func (s *dpdRenderersStub) Renderer(string) (render.Renderer, error) { return s.renderer, nil }

type dpdRendererStub struct {
	payload []byte
	result  model.LookupResult
}

func (s *dpdRendererStub) Format() string { return "markdown" }
func (s *dpdRendererStub) Render(_ context.Context, result model.LookupResult) ([]byte, error) {
	s.result = result
	return s.payload, nil
}

func TestModulePreservesJSONLookupSchema(t *testing.T) {
	lookup := &dpdLookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "solo", Format: "json"}, Entries: []model.Entry{{Headword: "solo", Content: "contenido"}}}}
	jsonBody, _ := json.Marshal(lookup.result)
	module := New(lookup, &dpdRenderersStub{renderer: &dpdRendererStub{payload: jsonBody}})

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "json", Sources: []string{"dpd"}})
	if err != nil {
		t.Fatalf("Execute() json error = %v", err)
	}
	if response.Fallback != nil {
		t.Fatalf("fallback = %#v, want nil on successful json response", response.Fallback)
	}
	if !strings.Contains(string(response.Body), `"Entries"`) {
		t.Fatalf("body = %s, want legacy lookup JSON schema", string(response.Body))
	}
}
