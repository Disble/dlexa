package render

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestSearchMarkdownRendererRendersOrderedCandidatesAndEmptyState(t *testing.T) {
	renderer := NewSearchMarkdownRenderer()
	result := model.SearchResult{Request: model.SearchRequest{Query: "abu dhabi"}, Candidates: []model.SearchCandidate{{DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi"}, {DisplayText: "⊗ alicuota", ArticleKey: "alícuoto"}}}

	payload, err := renderer.Render(context.Background(), result)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	text := string(payload)
	for _, want := range []string{"Candidate DPD entries for \"abu dhabi\":", "- Abu Dhabi -> Abu Dabi", "- ⊗ alicuota -> alícuoto"} {
		if !strings.Contains(text, want) {
			t.Fatalf("payload missing %q\n%s", want, text)
		}
	}

	emptyPayload, err := renderer.Render(context.Background(), model.SearchResult{Request: model.SearchRequest{Query: "zzz"}})
	if err != nil {
		t.Fatalf("Render() empty error = %v", err)
	}
	if !strings.Contains(string(emptyPayload), `No DPD entry candidates found for "zzz".`) {
		t.Fatalf("empty payload = %q", string(emptyPayload))
	}
}
