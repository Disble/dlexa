package render

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestRegistryReturnsRendererOrExplicitError(t *testing.T) {
	jsonRenderer := &testRenderer{format: "json"}
	registry := NewRegistry(jsonRenderer)

	resolved, err := registry.Renderer("json")
	if err != nil {
		t.Fatalf("Renderer(json) error = %v", err)
	}

	if resolved != jsonRenderer {
		t.Fatal("Renderer(json) did not return the registered renderer")
	}

	missing, err := registry.Renderer("xml")
	if err == nil {
		t.Fatal("Renderer(xml) error = nil, want explicit error")
	}

	if missing != nil {
		t.Fatal("Renderer(xml) returned a renderer for an unknown format")
	}

	if got := err.Error(); got != "renderer not registered for format \"xml\"" {
		t.Fatalf("Renderer(xml) error = %q, want %q", got, "renderer not registered for format \"xml\"")
	}
}

func TestRegistryResolvedRenderersPreserveStructuredMissContract(t *testing.T) {
	registry := NewRegistry(NewMarkdownRenderer(), NewJSONRenderer())
	result := model.LookupResult{
		Request: model.LookupRequest{Query: "zumbidoinexistente", Format: "markdown"},
		Misses: []model.LookupMiss{{
			Kind:   model.LookupMissKindGenericNotFound,
			Query:  "zumbidoinexistente",
			Source: "dpd",
			NextAction: &model.LookupNextAction{
				Kind:    model.LookupNextActionKindSearch,
				Query:   "zumbidoinexistente",
				Command: "dlexa search zumbidoinexistente",
			},
		}},
	}

	markdownRenderer, err := registry.Renderer("markdown")
	if err != nil {
		t.Fatalf("Renderer(markdown) error = %v", err)
	}
	markdownPayload, err := markdownRenderer.Render(context.Background(), result)
	if err != nil {
		t.Fatalf("markdown Render() error = %v", err)
	}
	if !strings.Contains(string(markdownPayload), "Try `dlexa search zumbidoinexistente`.") {
		t.Fatalf("markdown payload missing structured miss guidance\n%s", markdownPayload)
	}

	jsonRenderer, err := registry.Renderer("json")
	if err != nil {
		t.Fatalf("Renderer(json) error = %v", err)
	}
	jsonPayload, err := jsonRenderer.Render(context.Background(), result)
	if err != nil {
		t.Fatalf("json Render() error = %v", err)
	}

	var decoded struct {
		Misses []struct {
			Kind       string `json:"kind"`
			NextAction *struct {
				Command string `json:"command"`
			} `json:"next_action"`
		} `json:"misses"`
	}
	if err := json.Unmarshal(jsonPayload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(decoded.Misses) != 1 || decoded.Misses[0].Kind != string(model.LookupMissKindGenericNotFound) {
		t.Fatalf("decoded misses = %#v, want one generic miss", decoded.Misses)
	}
	if decoded.Misses[0].NextAction == nil || decoded.Misses[0].NextAction.Command != "dlexa search zumbidoinexistente" {
		t.Fatalf("decoded next action = %#v, want explicit search command", decoded.Misses[0].NextAction)
	}
}

type testRenderer struct {
	format string
}

func (r *testRenderer) Format() string {
	return r.format
}

func (r *testRenderer) Render(context.Context, model.LookupResult) ([]byte, error) {
	return nil, nil
}
