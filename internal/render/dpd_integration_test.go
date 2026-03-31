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

const renderErrFmt = "Render() error = %v"

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
	normalized, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, parsed)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	return normalized.Entries
}

func parseNormalizeDPDMiss(t *testing.T, term string) *model.LookupMiss {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", term+".html")) //nolint:gosec // G304: test fixture path from controlled input
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	parser := parse.NewDPDArticleParser()
	parsed, _, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd", DisplayName: "Diccionario panhispánico de dudas"}, fetch.Document{
		URL:         "https://www.rae.es/dpd/" + term,
		ContentType: "text/html; charset=utf-8",
		StatusCode:  200,
		Body:        body,
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	normalized, err := normalize.NewDPDNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, parsed)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if normalized.Miss == nil {
		t.Fatal("normalized.Miss = nil")
	}
	return normalized.Miss
}

func assertMarkdownMatchesGolden(t *testing.T, term string, entries []model.Entry) {
	t.Helper()
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: term, Format: "markdown"},
		Entries: entries,
	})
	if err != nil {
		t.Fatalf(renderErrFmt, err)
	}
	want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", term+".md.golden")) //nolint:gosec // G304: test fixture path from controlled input
	if err != nil {
		t.Fatalf("ReadFile() golden error = %v", err)
	}
	if stripANSITestOutput(string(payload)) != string(want) {
		t.Fatalf("Pipeline markdown mismatch\n--- got ---\n%s\n--- want ---\n%s", payload, want)
	}
}

func TestDPDParseNormalizeRenderMatchesBienGolden(t *testing.T) {
	entries := parseNormalizeDPD(t, "bien")
	assertMarkdownMatchesGolden(t, "bien", entries)
}

func TestDPDParseNormalizeRenderMatchesTildeGoldenAndJSONContract(t *testing.T) {
	entries := parseNormalizeDPD(t, "tilde")
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}

	assertMarkdownMatchesGolden(t, "tilde", entries)

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
		t.Fatalf(renderErrFmt, err)
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

func TestLookupMissMarkdownAndJSONStayInParity(t *testing.T) {
	tests := []struct {
		name        string
		result      model.LookupResult
		wantKind    string
		wantCommand string
		wantText    string
		wantDisplay string
	}{
		{
			name: "native suggestion parity",
			result: model.LookupResult{Request: model.LookupRequest{Query: "alicuota", Format: "markdown"}, Misses: []model.LookupMiss{{
				Kind:       model.LookupMissKindRelatedEntry,
				Query:      "alicuota",
				Source:     "dpd",
				Suggestion: &model.LookupSuggestion{Kind: "related_entry", DisplayText: "alícuota", URL: "https://www.rae.es/dpd/alícuota"},
			}}},
			wantKind:    "related_entry",
			wantText:    "Quizá quiso decir **alícuota**.",
			wantDisplay: "alícuota",
		},
		{
			name: "generic search nudge parity",
			result: model.LookupResult{Request: model.LookupRequest{Query: "zumbidoinexistente", Format: "markdown"}, Misses: []model.LookupMiss{{
				Kind:       model.LookupMissKindGenericNotFound,
				Query:      "zumbidoinexistente",
				Source:     "dpd",
				NextAction: &model.LookupNextAction{Kind: model.LookupNextActionKindSearch, Query: "zumbidoinexistente", Command: "dlexa search zumbidoinexistente"},
			}}},
			wantKind:    "generic_not_found",
			wantCommand: "dlexa search zumbidoinexistente",
			wantText:    "Try `dlexa search zumbidoinexistente`.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertLookupMissMarkdownParity(t, tt.result, tt.wantText)
			decoded := decodeLookupMissJSON(t, tt.result)
			assertDecodedLookupMiss(t, decoded, tt.wantKind, tt.wantCommand, tt.wantDisplay)
		})
	}
}

func assertLookupMissMarkdownParity(t *testing.T, result model.LookupResult, wantText string) {
	t.Helper()
	markdownPayload, err := NewMarkdownRenderer().Render(context.Background(), result)
	if err != nil {
		t.Fatalf(renderErrFmt, err)
	}
	if !strings.Contains(string(markdownPayload), wantText) {
		t.Fatalf("markdown missing %q\n%s", wantText, markdownPayload)
	}
}

func decodeLookupMissJSON(t *testing.T, result model.LookupResult) struct {
	Misses []struct {
		Kind       string `json:"kind"`
		Suggestion *struct {
			DisplayText string `json:"display_text"`
		} `json:"suggestion"`
		NextAction *struct {
			Command string `json:"command"`
		} `json:"next_action"`
	} `json:"misses"`
} {
	t.Helper()
	jsonPayload, err := NewJSONRenderer().Render(context.Background(), result)
	if err != nil {
		t.Fatalf("JSON Render() error = %v", err)
	}
	var decoded struct {
		Misses []struct {
			Kind       string `json:"kind"`
			Suggestion *struct {
				DisplayText string `json:"display_text"`
			} `json:"suggestion"`
			NextAction *struct {
				Command string `json:"command"`
			} `json:"next_action"`
		} `json:"misses"`
	}
	if err := json.Unmarshal(jsonPayload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return decoded
}

func assertDecodedLookupMiss(t *testing.T, decoded struct {
	Misses []struct {
		Kind       string `json:"kind"`
		Suggestion *struct {
			DisplayText string `json:"display_text"`
		} `json:"suggestion"`
		NextAction *struct {
			Command string `json:"command"`
		} `json:"next_action"`
	} `json:"misses"`
}, wantKind string, wantCommand string, wantDisplay string,
) {
	t.Helper()
	if len(decoded.Misses) != 1 || decoded.Misses[0].Kind != wantKind {
		t.Fatalf("decoded misses = %#v, want kind %q", decoded.Misses, wantKind)
	}
	if wantCommand != "" {
		if decoded.Misses[0].NextAction == nil || decoded.Misses[0].NextAction.Command != wantCommand {
			t.Fatalf("decoded next action = %#v, want %q", decoded.Misses[0].NextAction, wantCommand)
		}
		return
	}
	if decoded.Misses[0].Suggestion == nil || decoded.Misses[0].Suggestion.DisplayText != wantDisplay {
		t.Fatalf("decoded suggestion = %#v, want %q", decoded.Misses[0].Suggestion, wantDisplay)
	}
}

func TestGenericMissFixtureDoesNotRenderBogusSuggestionURL(t *testing.T) {
	miss := parseNormalizeDPDMiss(t, "zzzzz")
	if miss.Kind != model.LookupMissKindGenericNotFound {
		t.Fatalf("miss kind = %q, want %q", miss.Kind, model.LookupMissKindGenericNotFound)
	}
	if miss.Suggestion != nil {
		t.Fatalf("suggestion = %#v, want nil", miss.Suggestion)
	}
	payload, err := NewMarkdownRenderer().Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "zzzzz", Format: "markdown"},
		Misses:  []model.LookupMiss{*miss},
	})
	if err != nil {
		t.Fatalf(renderErrFmt, err)
	}
	text := string(payload)
	for _, forbidden := range []string{
		"Quizá quiso decir https://www.rae.es/dpd/",
		"https://www.rae.es/dpd/\n",
		"https://www.rae.es/dpd/.",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("markdown contains forbidden bogus suggestion %q\n%s", forbidden, text)
		}
	}
	if !strings.Contains(text, "Try `dlexa search zzzzz`.") {
		t.Fatalf("markdown missing explicit search nudge\n%s", text)
	}
}
