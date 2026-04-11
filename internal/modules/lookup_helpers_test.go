package modules

import (
	"context"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	modtest "github.com/Disble/dlexa/internal/modules/testsupport"
)

var buildLookupRequestCases = []struct {
	name          string
	request       Request
	defaultSource string
	want          model.LookupRequest
}{
	{
		name:          "trims fields and injects default source",
		request:       Request{Query: "  tilde  ", Format: " markdown ", NoCache: true},
		defaultSource: "duda-linguistica",
		want:          model.LookupRequest{Query: "tilde", Format: "markdown", Sources: []string{"duda-linguistica"}, NoCache: true},
	},
	{
		name:          "preserves provided sources",
		request:       Request{Query: "solo", Format: "json", Sources: []string{"espanol-al-dia"}},
		defaultSource: "ignored",
		want:          model.LookupRequest{Query: "solo", Format: "json", Sources: []string{"espanol-al-dia"}},
	},
}

func TestBuildLookupRequest(t *testing.T) {
	for _, tt := range buildLookupRequestCases {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildLookupRequest(tt.request, tt.defaultSource)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("BuildLookupRequest() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestExecuteLookupModuleUsesDefaultNotFoundFallback(t *testing.T) {
	lookup := &modtest.LookupStub{Result: model.LookupResult{Misses: []model.LookupMiss{{Kind: model.LookupMissKindGenericNotFound, Query: "solo"}}}}
	renderer := &modtest.RendererStub{FormatValue: "markdown", Payload: []byte("unused")}

	response, err := ExecuteLookupModule(
		context.Background(),
		Request{Query: "solo", Format: "markdown"},
		lookup,
		&modtest.RenderersStub{RendererValue: renderer},
		LookupModuleOptions{ModuleName: "espanol-al-dia", ModuleSource: "Español al día"},
	)
	if err != nil {
		t.Fatalf("ExecuteLookupModule() error = %v", err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindNotFound {
		t.Fatalf("fallback = %#v, want not_found fallback", response.Fallback)
	}
	if response.Fallback.NextCommand != "dlexa search solo" {
		t.Fatalf("next command = %q, want dlexa search solo", response.Fallback.NextCommand)
	}
	if got := lookup.Request.Sources; len(got) != 1 || got[0] != "espanol-al-dia" {
		t.Fatalf("lookup sources = %#v, want [\"espanol-al-dia\"]", got)
	}
}
