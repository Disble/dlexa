package render

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestSearchMarkdownRendererRendersOrderedCandidatesAndEmptyState(t *testing.T) {
	renderer := NewSearchMarkdownRenderer()
	result := model.SearchResult{Request: model.SearchRequest{Query: "abu dhabi"}, Candidates: []model.SearchCandidate{{Title: "Abu Dhabi", Snippet: "Entrada sugerida para la grafía recomendada.", NextCommand: "dlexa dpd Abu Dabi", Module: "dpd", ID: "Abu Dabi", DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi"}, {Title: "alícuoto", Snippet: "Artículo relacionado con la forma acentuada.", NextCommand: "dlexa dpd alícuoto", Module: "dpd", ID: "alícuoto", DisplayText: "⊗ alicuota", ArticleKey: "alícuoto"}}}

	payload, err := renderer.Render(context.Background(), result)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	text := string(payload)
	for _, want := range []string{"## Resultado semántico para \"abu dhabi\"", "### 1. Abu Dhabi", "- snippet: Entrada sugerida para la grafía recomendada.", "- next_command: `dlexa dpd Abu Dabi`", "### 2. alícuoto", "- snippet: Artículo relacionado con la forma acentuada.", "- next_command: `dlexa dpd alícuoto`"} {
		if !strings.Contains(text, want) {
			t.Fatalf("payload missing %q\n%s", want, text)
		}
	}

	emptyPayload, err := renderer.Render(context.Background(), model.SearchResult{Request: model.SearchRequest{Query: "zzz"}})
	if err != nil {
		t.Fatalf("Render() empty error = %v", err)
	}
	if !strings.Contains(string(emptyPayload), `No se encontraron rutas normativas accionables para "zzz".`) {
		t.Fatalf("empty payload = %q", string(emptyPayload))
	}
}
