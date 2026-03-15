package render

import (
	"context"
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

type testRenderer struct {
	format string
}

func (r *testRenderer) Format() string {
	return r.format
}

func (r *testRenderer) Render(context.Context, model.LookupResult) ([]byte, error) {
	return nil, nil
}
