package dudalinguistica

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	modtest "github.com/Disble/dlexa/internal/modules/testsupport"
)

const sourceDudaLinguistica = "duda-linguistica"

func TestModuleTargetsDudaLinguisticaSourceAndReturnsBody(t *testing.T) {
	lookup := &modtest.LookupStub{Result: model.LookupResult{Request: model.LookupRequest{Query: "tilde", Format: "markdown", Sources: []string{sourceDudaLinguistica}}, CacheHit: true, Entries: []model.Entry{{Headword: "tilde", Content: "contenido"}}}}
	renderers := &modtest.RenderersStub{RendererValue: &modtest.RendererStub{Payload: []byte("## Duda lingüística\ncontenido")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "tilde", Format: "markdown"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := lookup.Request.Sources; len(got) != 1 || got[0] != sourceDudaLinguistica {
		t.Fatalf("lookup sources = %#v, want [%q]", got, sourceDudaLinguistica)
	}
	if response.Source != "Duda lingüística" || string(response.Body) != "## Duda lingüística\ncontenido" {
		t.Fatalf("response = %#v", response)
	}
}

func TestModuleMapsNotFoundIntoStructuredFallback(t *testing.T) {
	lookup := &modtest.LookupStub{Result: model.LookupResult{Request: model.LookupRequest{Query: "tilde", Format: "markdown", Sources: []string{sourceDudaLinguistica}}, Misses: []model.LookupMiss{{Kind: model.LookupMissKindGenericNotFound, Query: "tilde"}}}}
	renderers := &modtest.RenderersStub{RendererValue: &modtest.RendererStub{Payload: []byte("unused")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "tilde", Format: "markdown", Sources: []string{sourceDudaLinguistica}})
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
