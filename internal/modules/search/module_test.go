package search

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
	if len(decoded.Candidates) != 4 {
		t.Fatalf("Candidates len = %d, want 4 after filtering institutional noise", len(decoded.Candidates))
	}

	assertCandidate(t, decoded.Candidates[0], "dpd", "bien", "dlexa dpd bien")
	assertCandidate(t, decoded.Candidates[1], "noticia", "preguntas-frecuentes-sobre-la-tilde", "dlexa noticia preguntas-frecuentes-sobre-la-tilde")
	if decoded.Candidates[1].Classification != "faq" {
		t.Fatalf("faq candidate classification = %q, want faq", decoded.Candidates[1].Classification)
	}
	if decoded.Candidates[1].Deferred {
		t.Fatalf("noticia FAQ candidate should be executable, got deferred=%v", decoded.Candidates[1].Deferred)
	}
	assertCandidate(t, decoded.Candidates[2], "espanol-al-dia", "la-conjuncion-o-siempre-sin-tilde", "dlexa espanol-al-dia la-conjuncion-o-siempre-sin-tilde")
	if decoded.Candidates[2].Deferred {
		t.Fatalf("espanol-al-dia candidate should be executable, got deferred=%v", decoded.Candidates[2].Deferred)
	}
	if decoded.Candidates[3].NextCommand == "" || !strings.Contains(decoded.Candidates[3].NextCommand, "dlexa search") {
		t.Fatalf("unknown candidate next command = %q, want safe fallback", decoded.Candidates[3].NextCommand)
	}
	if decoded.Candidates[3].Module != "unknown" {
		t.Fatalf("unknown candidate module = %q, want unknown", decoded.Candidates[3].Module)
	}
}

func TestModuleReturnsExplicitNoResultsWhenCuratedResultsAreEmpty(t *testing.T) {
	searcher := &searchStub{result: model.SearchResult{Request: model.SearchRequest{Query: "zzz", Format: "markdown"}, Candidates: []model.SearchCandidate{{Title: "Institucional", URL: "https://www.rae.es/institucion/discurso"}}}}
	renderer := &searchRendererStub{payload: []byte("sin resultados")}
	module := New(searcher, &searchRenderersStub{renderer: renderer})

	response, err := module.Execute(context.Background(), modules.Request{Query: "zzz", Format: "markdown"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback != nil {
		t.Fatalf("fallback = %#v, want nil", response.Fallback)
	}
	if got := renderer.result.Outcome; got != model.SearchOutcomeNoResults {
		t.Fatalf("Outcome = %q, want %q", got, model.SearchOutcomeNoResults)
	}
	if string(response.Body) != "sin resultados" {
		t.Fatalf("Body = %q, want renderer payload", string(response.Body))
	}
}

func TestModuleKeepsFailuresOnExplicitFallbackPath(t *testing.T) {
	searcher := &searchStub{err: model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchParseFailed, Message: "markup changed", Source: "search", Severity: model.ProblemSeverityError}, nil)}
	module := New(searcher, &searchRenderersStub{renderer: &searchRendererStub{payload: []byte("unused")}})

	response, err := module.Execute(context.Background(), modules.Request{Query: "zzz", Format: "markdown"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindParseFailure {
		t.Fatalf("Fallback = %#v, want parse failure", response.Fallback)
	}
}

func TestModuleMapsRateLimitedFailuresToDedicatedFallback(t *testing.T) {
	searcher := &searchStub{err: model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchRateLimited, Message: "limited", Source: "search", Severity: model.ProblemSeverityError}, nil)}
	module := New(searcher, &searchRenderersStub{renderer: &searchRendererStub{payload: []byte("unused")}})

	response, err := module.Execute(context.Background(), modules.Request{Query: "zzz", Format: "markdown"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindRateLimited {
		t.Fatalf("Fallback = %#v, want rate limited", response.Fallback)
	}
}

func TestModuleForwardsExplicitSourcesToSearcher(t *testing.T) {
	searcher := &searchStub{result: model.SearchResult{Request: model.SearchRequest{Query: "tilde", Format: "json"}}}
	module := New(searcher, &searchRenderersStub{renderer: &searchRendererStub{payload: []byte("{}")}})
	request := modules.Request{Query: "tilde", Format: "json", Sources: []string{"search", "academia"}}

	if _, err := module.Execute(context.Background(), request); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !reflect.DeepEqual(searcher.request.Sources, request.Sources) {
		t.Fatalf("searcher request sources = %#v, want %#v", searcher.request.Sources, request.Sources)
	}
}

func TestModuleDerivesSourceFromRequestedProviders(t *testing.T) {
	tests := []struct {
		name        string
		request     modules.Request
		searcherReq model.SearchRequest
		wantSource  string
	}{
		{
			name:        "dpd only search uses dpd label",
			request:     modules.Request{Query: "solo", Format: "markdown", Sources: []string{"dpd"}},
			searcherReq: model.SearchRequest{Query: "solo", Format: "markdown", Sources: []string{"dpd"}},
			wantSource:  "Diccionario panhispánico de dudas",
		},
		{
			name:        "general search keeps general label",
			request:     modules.Request{Query: "solo", Format: "markdown", Sources: []string{"search"}},
			searcherReq: model.SearchRequest{Query: "solo", Format: "markdown", Sources: []string{"search"}},
			wantSource:  "búsqueda general RAE",
		},
		{
			name:        "multiple sources use federated general label",
			request:     modules.Request{Query: "solo", Format: "markdown", Sources: []string{"search", "dpd"}},
			searcherReq: model.SearchRequest{Query: "solo", Format: "markdown", Sources: []string{"search", "dpd"}},
			wantSource:  "búsqueda general RAE",
		},
		{
			name:        "unknown single source passes through",
			request:     modules.Request{Query: "solo", Format: "markdown", Sources: []string{"academia"}},
			searcherReq: model.SearchRequest{Query: "solo", Format: "markdown", Sources: []string{"academia"}},
			wantSource:  "academia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searcher := &searchStub{result: model.SearchResult{Request: tt.searcherReq}}
			module := New(searcher, &searchRenderersStub{renderer: &searchRendererStub{payload: []byte("ok")}})

			response, err := module.Execute(context.Background(), tt.request)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if response.Source != tt.wantSource {
				t.Fatalf("Source = %q, want %q", response.Source, tt.wantSource)
			}
		})
	}
}

func TestModuleDerivesFallbackSourceFromRequestedProviders(t *testing.T) {
	searcher := &searchStub{err: model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "dpd unavailable", Severity: model.ProblemSeverityError}, nil)}
	module := New(searcher, &searchRenderersStub{renderer: &searchRendererStub{payload: []byte("unused")}})

	response, err := module.Execute(context.Background(), modules.Request{Query: "solo", Format: "markdown", Sources: []string{"dpd"}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Source != "Diccionario panhispánico de dudas" {
		t.Fatalf("Source = %q, want DPD label on fallback path", response.Source)
	}
	if response.Fallback == nil {
		t.Fatal("Fallback = nil, want structured fallback")
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
