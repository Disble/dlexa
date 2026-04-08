package render

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestSearchMarkdownRendererRendersOrderedCandidatesAndEmptyState(t *testing.T) {
	renderer := NewSearchMarkdownRenderer()
	result := model.SearchResult{Request: model.SearchRequest{Query: "abu dhabi"}, Outcome: model.SearchOutcomeResults, Candidates: []model.SearchCandidate{{Title: "Abu Dhabi", Snippet: "Entrada sugerida para la grafía recomendada.", NextCommand: "dlexa dpd Abu Dabi", Module: "dpd", ID: "Abu Dabi", DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi"}, {Title: "Tilde en solo", Snippet: "Artículo de orientación adicional.", NextCommand: "dlexa espanol-al-dia solo", Module: "espanol-al-dia", ID: "solo", Deferred: true, DisplayText: "Tilde en solo"}, {Title: "Ruta rara", Snippet: "Resultado visible pero no mapeado.", NextCommand: "dlexa search abu dhabi", Module: "unknown", ID: "ruta-rara", DisplayText: "Ruta rara", URL: "https://www.rae.es/archivo/ruta-rara"}}}

	payload, err := renderer.Render(context.Background(), result)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	text := string(payload)
	for _, want := range []string{"## Resultado semántico para \"abu dhabi\"", "### 1. Abu Dhabi", "- sugerencia: `dlexa dpd Abu Dabi`", "### 2. Tilde en solo", "- More info: `dlexa espanol-al-dia solo`", "- nota: (not yet available as CLI command)", "### 3. Ruta rara", "- url: https://www.rae.es/archivo/ruta-rara", "- fallback_command: `dlexa search abu dhabi`"} {
		if !strings.Contains(text, want) {
			t.Fatalf("payload missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "- sugerencia: `dlexa espanol-al-dia solo`") {
		t.Fatalf("deferred candidate rendered as executable suggestion\n%s", text)
	}

	emptyPayload, err := renderer.Render(context.Background(), model.SearchResult{Request: model.SearchRequest{Query: "zzz"}, Outcome: model.SearchOutcomeNoResults})
	if err != nil {
		t.Fatalf("Render() empty error = %v", err)
	}
	if !strings.Contains(string(emptyPayload), `No se encontraron resultados normativos útiles para "zzz".`) {
		t.Fatalf("empty payload = %q", string(emptyPayload))
	}
}
