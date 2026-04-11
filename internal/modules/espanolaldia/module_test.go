package espanolaldia

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	modtest "github.com/Disble/dlexa/internal/modules/testsupport"
)

const (
	sourceEspanolAlDia = "espanol-al-dia"
	executeErrFormat   = "Execute() error = %v"
)

func TestModuleTargetsEspanolAlDiaSourceAndReturnsBody(t *testing.T) {
	lookup := &modtest.LookupStub{Result: model.LookupResult{Request: model.LookupRequest{Query: "solo", Format: "markdown", Sources: []string{sourceEspanolAlDia}}, CacheHit: true, Entries: []model.Entry{{Headword: "solo", Content: "contenido"}}}}
	renderers := &modtest.RenderersStub{RendererValue: &modtest.RendererStub{Payload: []byte("## Español al día\ncontenido")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "markdown"})
	if err != nil {
		t.Fatalf(executeErrFormat, err)
	}
	if got := lookup.Request.Sources; len(got) != 1 || got[0] != sourceEspanolAlDia {
		t.Fatalf("lookup sources = %#v, want [%q]", got, sourceEspanolAlDia)
	}
	if response.Source != "Español al día" || string(response.Body) != "## Español al día\ncontenido" {
		t.Fatalf("response = %#v", response)
	}
}

func TestModuleMapsNotFoundIntoStructuredFallback(t *testing.T) {
	lookup := &modtest.LookupStub{Result: model.LookupResult{Request: model.LookupRequest{Query: "solo", Format: "json", Sources: []string{sourceEspanolAlDia}}, Misses: []model.LookupMiss{{Kind: model.LookupMissKindGenericNotFound, Query: "solo"}}}}
	renderers := &modtest.RenderersStub{RendererValue: &modtest.RendererStub{Payload: []byte("{}")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "json", Sources: []string{sourceEspanolAlDia}})
	if err != nil {
		t.Fatalf(executeErrFormat, err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindNotFound {
		t.Fatalf("fallback = %#v, want not_found fallback", response.Fallback)
	}
	if response.Fallback.NextCommand != "dlexa search solo" {
		t.Fatalf("next command = %q, want dlexa search solo", response.Fallback.NextCommand)
	}
}

func TestModulePreservesJSONLookupSchema(t *testing.T) {
	lookup := &modtest.LookupStub{Result: model.LookupResult{Request: model.LookupRequest{Query: "solo", Format: "json", Sources: []string{sourceEspanolAlDia}}, Entries: []model.Entry{{Headword: "solo", Content: "contenido"}}}}
	jsonBody, _ := json.Marshal(lookup.Result)
	module := New(lookup, &modtest.RenderersStub{RendererValue: &modtest.RendererStub{Payload: jsonBody}})

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "json", Sources: []string{sourceEspanolAlDia}})
	if err != nil {
		t.Fatalf(executeErrFormat, err)
	}
	if response.Fallback != nil {
		t.Fatalf("fallback = %#v, want nil", response.Fallback)
	}
	if !strings.Contains(string(response.Body), `"Entries"`) {
		t.Fatalf("body = %s, want lookup JSON schema", string(response.Body))
	}
}
