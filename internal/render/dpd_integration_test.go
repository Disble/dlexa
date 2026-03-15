package render

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
)

func parseNormalizeDPD(t *testing.T, term string) []model.Entry {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", term+".html")) //nolint:gosec // G304: test fixture path from controlled input
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	parser := parse.NewDPDArticleParser()
	parsed, _, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd", DisplayName: term}, fetch.Document{
		URL:         "https://www.rae.es/dpd/" + term,
		ContentType: "text/html; charset=utf-8",
		StatusCode:  200,
		Body:        body,
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	normalizer := normalize.NewDPDNormalizer()
	entries, _, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, parsed)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	return entries
}

func TestDPDParseNormalizeRenderMatchesBienGolden(t *testing.T) {
	entries := parseNormalizeDPD(t, "bien")
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "bien", Format: "markdown"},
		Entries: entries,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", "bien.md.golden"))
	if err != nil {
		t.Fatalf("ReadFile() golden error = %v", err)
	}

	if got := stripANSITestOutput(string(payload)); got != string(want) {
		t.Fatalf("Pipeline markdown mismatch\n--- got ---\n%s\n--- want ---\n%s", payload, want)
	}
}

func TestDPDParseNormalizeRenderMatchesTildeGoldenAndJSONContract(t *testing.T) {
	entries := parseNormalizeDPD(t, "tilde")
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}

	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "tilde", Format: "markdown"},
		Entries: entries,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", "tilde.md.golden"))
	if err != nil {
		t.Fatalf("ReadFile() golden error = %v", err)
	}
	if got := stripANSITestOutput(string(payload)); got != string(want) {
		t.Fatalf("Pipeline markdown mismatch\n--- got ---\n%s\n--- want ---\n%s", payload, want)
	}

	jsonRenderer := NewJSONRenderer()
	jsonPayload, err := jsonRenderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "tilde", Format: "json"},
		Entries: entries,
	})
	if err != nil {
		t.Fatalf("JSON Render() error = %v", err)
	}
	jsonWant, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", "tilde.json.golden"))
	if err != nil {
		t.Fatalf("ReadFile() json golden error = %v", err)
	}
	var gotJSON any
	var wantJSON any
	if err := json.Unmarshal(jsonPayload, &gotJSON); err != nil {
		t.Fatalf("json.Unmarshal(got) error = %v", err)
	}
	if err := json.Unmarshal(jsonWant, &wantJSON); err != nil {
		t.Fatalf("json.Unmarshal(want) error = %v", err)
	}
	if !reflect.DeepEqual(gotJSON, wantJSON) {
		t.Fatalf("JSON mismatch\n--- got ---\n%s\n--- want ---\n%s", jsonPayload, jsonWant)
	}
}

func TestDPDParseNormalizeRenderProducesSemanticMarkdownOutput(t *testing.T) {
	entries := parseNormalizeDPD(t, "bien")
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "bien", Format: "markdown"},
		Entries: entries,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	text := string(payload)
	for _, raw := range []string{"[ej.:", "ej.:", "‹", "›", "\x1b["} {
		if strings.Contains(text, raw) {
			t.Fatalf("markdown output still contains rejected raw marker %q\n%s", raw, text)
		}
	}
	for _, want := range []string{
		"*Cierra bien la ventana, por favor*",
		"*No he dormido bien esta noche*",
		"5. bien que. Locución conjuntiva concesiva equivalente a 'aunque'. Con este mismo sentido, se emplea más frecuentemente la locución",
		"6. más bien. Locución adverbial que se usa con distintos valores:",
		"  a) Para introducir una rectificación o una matización.",
		"  b) Con el sentido de 'en cierto modo, de alguna manera'.",
		"  c) También significa 'mejor o preferentemente'.",
		"→ [6](bien#S1590507271213267522)",
		"→ [7](bien#S1590507271244936818)",
		"*mejor*",
		"*más bien*",
		"*si bien*",
		"*bien que*",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("markdown output missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "'satisfactoriamente': No he dormido bien esta noche.") {
		t.Fatalf("markdown output collapsed real bien example into plain prose\n%s", text)
	}
	for _, want := range []string{"El comparativo es *mejor*", "No debe usarse *más bien* como comparativo", "la locución *si bien*", "la locución *bien que*"} {
		if !strings.Contains(text, want) {
			t.Fatalf("markdown output missing authored semantic text %q\n%s", want, text)
		}
	}
	for _, forbidden := range []string{"→ 6", "→ 7"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("markdown output contains forbidden plain-text projection %q\n%s", forbidden, text)
		}
	}
}
