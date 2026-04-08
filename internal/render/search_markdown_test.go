package render

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

var (
	searchMarkdownOrderedCandidates = []model.SearchCandidate{
		{Title: "Abu Dhabi", Snippet: "Entrada sugerida para la grafía recomendada.", NextCommand: "dlexa dpd Abu Dabi", Module: "dpd", ID: "Abu Dabi", DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi", SourceHint: "Diccionario panhispánico de dudas"},
		{Title: "Tilde en solo", Snippet: "Artículo de orientación adicional.", NextCommand: "dlexa espanol-al-dia solo", Module: "espanol-al-dia", ID: "solo", Deferred: true, DisplayText: "Tilde en solo", SourceHint: "Academia"},
		{Title: "Ruta rara", Snippet: "Resultado visible pero no mapeado.", NextCommand: "dlexa search abu dhabi", Module: "unknown", ID: "ruta-rara", DisplayText: "Ruta rara", URL: "https://www.rae.es/archivo/ruta-rara", SourceHint: "Archivo externo"},
	}
	searchMarkdownDeferredWithURLCandidate    = model.SearchCandidate{Title: "Uso del guion", Deferred: true, URL: "https://www.rae.es/espanol-al-dia/uso-del-guion", NextCommand: "dlexa espanol-al-dia uso-del-guion", Classification: "linguistic-article"}
	searchMarkdownDeferredWithoutURLCandidate = model.SearchCandidate{Title: "Sin URL", Deferred: true, URL: "", NextCommand: "dlexa espanol-al-dia sin-url"}
	searchMarkdownNonDeferredCandidate        = model.SearchCandidate{Title: "solo", Deferred: false, NextCommand: "dlexa dpd solo", Module: "dpd"}
)

func TestSearchMarkdownRendererRendersOrderedCandidatesAndEmptyState(t *testing.T) {
	renderer := NewSearchMarkdownRenderer()
	result := model.SearchResult{Request: model.SearchRequest{Query: "abu dhabi"}, Outcome: model.SearchOutcomeResults, Candidates: searchMarkdownOrderedCandidates, Problems: []model.Problem{{Source: "academia", Message: "timeout", Severity: model.ProblemSeverityError}}}

	payload, err := renderer.Render(context.Background(), result)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	text := string(payload)
	for _, want := range []string{"## Resultado semántico para \"abu dhabi\"", "### 1. Abu Dhabi", "- fuente: Diccionario panhispánico de dudas", "- sugerencia: `dlexa dpd Abu Dabi`", "### 2. Tilde en solo", "- fuente: Academia", "_(Acceso futuro via CLI: dlexa espanol-al-dia solo)_", "### 3. Ruta rara", "- fuente: Archivo externo", "- url: https://www.rae.es/archivo/ruta-rara", "- fallback_command: `dlexa search abu dhabi`", "## Problemas detectados", "- [academia] timeout"} {
		if !strings.Contains(text, want) {
			t.Fatalf("payload missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "- sugerencia: `dlexa espanol-al-dia solo`") {
		t.Fatalf("deferred candidate rendered as executable suggestion\n%s", text)
	}
	if strings.Contains(text, "More info:") || strings.Contains(text, "not yet available as CLI command") {
		t.Fatalf("legacy deferred guidance remained\n%s", text)
	}

	emptyPayload, err := renderer.Render(context.Background(), model.SearchResult{Request: model.SearchRequest{Query: "zzz"}, Outcome: model.SearchOutcomeNoResults, Problems: []model.Problem{{Source: "search", Message: "upstream degraded", Severity: model.ProblemSeverityError}}})
	if err != nil {
		t.Fatalf("Render() empty error = %v", err)
	}
	if !strings.Contains(string(emptyPayload), `No se encontraron resultados normativos útiles para "zzz".`) {
		t.Fatalf("empty payload = %q", string(emptyPayload))
	}
	if !strings.Contains(string(emptyPayload), "- [search] upstream degraded") {
		t.Fatalf("empty payload missing problems = %q", string(emptyPayload))
	}
}

func TestSearchMarkdownDeferredAccessAndWarnings(t *testing.T) {
	renderer := NewSearchMarkdownRenderer()
	tests := []struct {
		name        string
		result      model.SearchResult
		contains    []string
		notContains []string
	}{
		{
			name: "deferred with url",
			result: model.SearchResult{
				Request:    model.SearchRequest{Query: "guion"},
				Outcome:    model.SearchOutcomeResults,
				Candidates: []model.SearchCandidate{searchMarkdownDeferredWithURLCandidate},
			},
			contains: []string{"🌐 https://www.rae.es/espanol-al-dia/uso-del-guion"},
		},
		{
			name: "deferred without url",
			result: model.SearchResult{
				Request:    model.SearchRequest{Query: "sin url"},
				Outcome:    model.SearchOutcomeResults,
				Candidates: []model.SearchCandidate{searchMarkdownDeferredWithoutURLCandidate},
			},
			notContains: []string{"🌐"},
		},
		{
			name: "deferred with next command",
			result: model.SearchResult{
				Request:    model.SearchRequest{Query: "guion"},
				Outcome:    model.SearchOutcomeResults,
				Candidates: []model.SearchCandidate{searchMarkdownDeferredWithURLCandidate},
			},
			contains:    []string{"Acceso futuro via CLI"},
			notContains: []string{"More info:", "not yet available as CLI command"},
		},
		{
			name: "non-deferred unchanged",
			result: model.SearchResult{
				Request:    model.SearchRequest{Query: "solo"},
				Outcome:    model.SearchOutcomeResults,
				Candidates: []model.SearchCandidate{searchMarkdownNonDeferredCandidate},
			},
			contains:    []string{"- sugerencia: `dlexa dpd solo`"},
			notContains: []string{"🌐", "Acceso futuro"},
		},
		{
			name: "warnings present",
			result: model.SearchResult{
				Request:    model.SearchRequest{Query: "advertencias"},
				Outcome:    model.SearchOutcomeResults,
				Candidates: []model.SearchCandidate{searchMarkdownNonDeferredCandidate},
				Warnings: []model.Warning{
					{Message: "Fuente parcialmente degradada"},
					{Message: "Caché desactualizada"},
				},
			},
			contains: []string{"⚠️", "Fuente parcialmente degradada", "Caché desactualizada"},
		},
		{
			name: "warnings empty",
			result: model.SearchResult{
				Request:    model.SearchRequest{Query: "sin advertencias"},
				Outcome:    model.SearchOutcomeResults,
				Candidates: []model.SearchCandidate{searchMarkdownNonDeferredCandidate},
				Warnings:   nil,
			},
			notContains: []string{"⚠️"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := renderer.Render(context.Background(), tt.result)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			text := string(payload)
			for _, want := range tt.contains {
				if !strings.Contains(text, want) {
					t.Fatalf("payload missing %q\n%s", want, text)
				}
			}
			for _, want := range tt.notContains {
				if strings.Contains(text, want) {
					t.Fatalf("payload unexpectedly contained %q\n%s", want, text)
				}
			}
		})
	}
}
