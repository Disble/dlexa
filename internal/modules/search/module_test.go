package search

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/render"
)

func TestModuleFiltersNoiseRescuesFAQAndMapsCommands(t *testing.T) {
	fixture := loadFixture(t)
	searcher := &searchStub{result: model.SearchResult{Request: model.SearchRequest{Query: "solo o sólo", Format: "json"}, Candidates: fixture}}
	renderer := &searchRendererStub{}
	module := New(searcher, &searchRenderersStub{renderer: renderer})

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo o sólo", Format: "json"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback != nil {
		t.Fatalf("fallback = %#v, want nil", response.Fallback)
	}

	decoded := renderer.result
	if len(decoded.Candidates) != 3 {
		t.Fatalf("Candidates len = %d, want 3 after filtering institutional noise", len(decoded.Candidates))
	}

	assertCandidate(t, decoded.Candidates[0], "espanol-al-dia", "la-conjuncion-o-siempre-sin-tilde", "dlexa espanol-al-dia la-conjuncion-o-siempre-sin-tilde")
	assertCandidate(t, decoded.Candidates[1], "noticia", "preguntas-frecuentes-sobre-la-tilde", "dlexa noticia preguntas-frecuentes-sobre-la-tilde")
	if decoded.Candidates[1].Classification != "faq" {
		t.Fatalf("faq candidate classification = %q, want faq", decoded.Candidates[1].Classification)
	}
	if decoded.Candidates[2].NextCommand == "" || !strings.Contains(decoded.Candidates[2].NextCommand, "dlexa search") {
		t.Fatalf("unknown candidate next command = %q, want safe fallback", decoded.Candidates[2].NextCommand)
	}
	if decoded.Candidates[2].Module != "unknown" {
		t.Fatalf("unknown candidate module = %q, want unknown", decoded.Candidates[2].Module)
	}
}

func TestModuleReturnsNotFoundFallbackWhenCuratedResultsAreEmpty(t *testing.T) {
	searcher := &searchStub{result: model.SearchResult{Request: model.SearchRequest{Query: "zzz", Format: "markdown"}, Candidates: []model.SearchCandidate{{Title: "Institucional", URL: "https://www.rae.es/institucion/discurso"}}}}
	module := New(searcher, &searchRenderersStub{renderer: &searchRendererStub{payload: []byte("unused")}})

	response, err := module.Execute(context.Background(), modules.Request{Query: "zzz", Format: "markdown"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindNotFound || response.Fallback.NextCommand != "dlexa search zzz" {
		t.Fatalf("fallback = %#v", response.Fallback)
	}
}

func loadFixture(t *testing.T) []model.SearchCandidate {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "candidates.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var candidates []model.SearchCandidate
	if err := json.Unmarshal(raw, &candidates); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return candidates
}

func assertCandidate(t *testing.T, candidate model.SearchCandidate, wantModule, wantID, wantCommand string) {
	t.Helper()
	if candidate.Module != wantModule || candidate.ID != wantID || candidate.NextCommand != wantCommand {
		t.Fatalf("candidate = %#v", candidate)
	}
}

type searchStub struct {
	request model.SearchRequest
	result  model.SearchResult
	err     error
}

func (s *searchStub) Search(_ context.Context, request model.SearchRequest) (model.SearchResult, error) {
	s.request = request
	return s.result, s.err
}

type searchRenderersStub struct{ renderer render.SearchRenderer }

func (s *searchRenderersStub) Renderer(string) (render.SearchRenderer, error) { return s.renderer, nil }

type searchRendererStub struct {
	payload []byte
	result  model.SearchResult
}

func (s *searchRendererStub) Format() string { return "json" }
func (s *searchRendererStub) Render(_ context.Context, result model.SearchResult) ([]byte, error) {
	s.result = result
	if s.payload != nil {
		return s.payload, nil
	}
	return json.Marshal(result)
}
