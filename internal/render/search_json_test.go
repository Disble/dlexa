package render

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestSearchJSONRendererPreservesRawHTMLAndOrderWithoutMarkdownProjection(t *testing.T) {
	renderer := NewSearchJSONRenderer()
	result := model.SearchResult{Request: model.SearchRequest{Query: "abu dhabi", Format: "json"}, Outcome: model.SearchOutcomeNoResults, Candidates: []model.SearchCandidate{{RawLabelHTML: `<em>Abu Dhabi</em>`, DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi"}, {RawLabelHTML: `<span class="bolaspa">⊗</span>alicuota`, DisplayText: "⊗ alicuota", ArticleKey: "alícuoto", Module: "unknown", URL: "https://www.rae.es/archivo/ruta-rara", NextCommand: "dlexa search abu dhabi"}, {RawLabelHTML: `<strong>solo</strong>`, DisplayText: "solo", Module: "espanol-al-dia", NextCommand: "dlexa espanol-al-dia solo", Deferred: true}}}

	payload, err := renderer.Render(context.Background(), result)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	text := string(payload)
	if !strings.Contains(text, `<em>Abu Dhabi</em>`) {
		t.Fatalf("payload = %s, want raw html unescaped", text)
	}
	if !strings.Contains(text, `"Request": {`) || !strings.Contains(text, `"Candidates": [`) || !strings.Contains(text, `"Outcome": "no_results"`) {
		t.Fatalf("payload = %s, want current top-level Go field names", text)
	}
	if strings.Contains(text, "->") {
		t.Fatalf("payload = %s, json must not inherit markdown projection tokens", text)
	}
	if !strings.Contains(text, `"deferred": true`) {
		t.Fatalf("payload = %s, want deferred field for deferred candidates", text)
	}
	if !strings.Contains(text, `"deferred": false`) {
		t.Fatalf("payload = %s, want deferred field for non-deferred candidates too", text)
	}

	var decoded model.SearchResult
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(decoded.Candidates) != 3 || decoded.Candidates[0].ArticleKey != "Abu Dabi" || decoded.Candidates[1].ArticleKey != "alícuoto" || !decoded.Candidates[2].Deferred {
		t.Fatalf("decoded = %#v", decoded)
	}
}
