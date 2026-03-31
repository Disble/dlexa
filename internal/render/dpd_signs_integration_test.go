package render

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
)

type dpdFixtureOutput struct {
	entries []model.Entry
	md      string
	json    []byte
}

type dpdSignGoldenSnapshot struct {
	Headword    string   `json:"headword"`
	Markdown    string   `json:"markdown"`
	InlineKinds []string `json:"inline_kinds"`
}

const (
	termAbrogar        = "abrogar"
	termAcertar        = "acertar"
	termAlicuota       = "alícuota"
	termAndrofobia     = "androfobia"
	dpdTestBaseURL     = "https://www.rae.es/dpd/"
	dpdHTMLContentType = "text/html; charset=utf-8"
	exclusionMarker    = "⊗"
	acertarReference   = "→ [acertar](/dpd/ayuda/modelos-de-conjugacion-verbal#acertar)"
)

func dpdTestArticleURL(term string) string {
	return dpdTestBaseURL + term
}

func parseNormalizeDPDEntries(t *testing.T, term string, body []byte) []model.Entry {
	t.Helper()

	parser := parse.NewDPDArticleParser()
	parsed, _, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd", DisplayName: term}, fetch.Document{
		URL:         dpdTestArticleURL(term),
		ContentType: dpdHTMLContentType,
		StatusCode:  http.StatusOK,
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

func dpdFixturePath(name string) string {
	return filepath.Join("..", "parse", "testdata", name+".html")
}

func dpdGoldenPath(name, extension string) string {
	return filepath.Join("..", "parse", "testdata", name+".golden."+extension)
}

func parseNormalizeDPDFixtureFromPath(t *testing.T, term string, fixturePath string) []model.Entry {
	t.Helper()
	body, err := os.ReadFile(fixturePath) //nolint:gosec // G304: controlled test fixture path
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", fixturePath, err)
	}
	return parseNormalizeDPDEntries(t, term, body)
}

func renderDPDFixtureOutputs(t *testing.T, term string, fixturePath string) dpdFixtureOutput {
	t.Helper()
	entries := parseNormalizeDPDFixtureFromPath(t, term, fixturePath)
	return renderDPDEntries(t, term, entries)
}

func renderDPDEntries(t *testing.T, term string, entries []model.Entry) dpdFixtureOutput {
	t.Helper()

	markdownRenderer := NewMarkdownRenderer()
	markdownPayload, err := markdownRenderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: term, Format: "markdown"},
		Entries: entries,
	})
	if err != nil {
		t.Fatalf("markdown Render() error = %v", err)
	}

	jsonRenderer := NewJSONRenderer()
	jsonPayload, err := jsonRenderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: term, Format: "json"},
		Entries: entries,
	})
	if err != nil {
		t.Fatalf("json Render() error = %v", err)
	}

	return dpdFixtureOutput{
		entries: entries,
		md:      stripANSITestOutput(string(markdownPayload)),
		json:    jsonPayload,
	}
}

func renderDPDRawHTMLOutputs(t *testing.T, term string, raw string) dpdFixtureOutput {
	t.Helper()
	return renderDPDEntries(t, term, parseNormalizeDPDEntries(t, term, []byte(raw)))
}

func requireJSONContainsInlineKind(t *testing.T, payload []byte, want string) {
	t.Helper()
	if !strings.Contains(string(payload), `"Kind": "`+want+`"`) {
		t.Fatalf("json output missing inline kind %q\n%s", want, payload)
	}
}

func requireMarkdownContainsPlainBracket(t *testing.T, payload string, want string) {
	t.Helper()
	if !strings.Contains(payload, want) {
		t.Fatalf("markdown output missing bracket text %q\n%s", want, payload)
	}
	for _, forbidden := range []string{"<dfn>", `<span class="nn">`, `<span class="yy">`} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("markdown output leaked raw semantic html %q\n%s", forbidden, payload)
		}
	}
}

func requireMarkdownGolden(t *testing.T, got string, goldenPath string) {
	t.Helper()
	want, err := os.ReadFile(goldenPath) //nolint:gosec // G304: controlled test fixture path
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v\n--- got markdown ---\n%s", goldenPath, err, got)
	}
	if got != string(want) {
		t.Fatalf("markdown mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", goldenPath, got, want)
	}
}

func requireJSONGolden(t *testing.T, got []byte, goldenPath string) {
	t.Helper()
	want, err := os.ReadFile(goldenPath) //nolint:gosec // G304: controlled test fixture path
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v\n--- got json ---\n%s", goldenPath, err, got)
	}
	var gotJSON dpdSignGoldenSnapshot
	var wantJSON dpdSignGoldenSnapshot
	if err := json.Unmarshal(got, &gotJSON); err != nil {
		t.Fatalf("json.Unmarshal(got) error = %v", err)
	}
	if err := json.Unmarshal(want, &wantJSON); err != nil {
		t.Fatalf("json.Unmarshal(want) error = %v", err)
	}
	if !reflect.DeepEqual(gotJSON, wantJSON) {
		t.Fatalf("json mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", goldenPath, got, want)
	}
}

func collectInlineKinds(inlines []model.Inline, seen map[string]struct{}) {
	for _, inline := range inlines {
		if inline.Kind != model.InlineKindText {
			seen[inline.Kind] = struct{}{}
		}
		collectInlineKinds(inline.Children, seen)
	}
}

func buildSignGoldenSnapshot(entries []model.Entry) ([]byte, error) {
	if len(entries) == 0 || entries[0].Article == nil {
		return json.MarshalIndent(dpdSignGoldenSnapshot{}, "", "  ")
	}

	seen := make(map[string]struct{})
	for _, section := range entries[0].Article.Sections {
		for _, block := range section.Blocks {
			if block.Paragraph == nil {
				continue
			}
			collectInlineKinds(block.Paragraph.Inlines, seen)
		}
	}

	kinds := make([]string, 0, len(seen))
	for kind := range seen {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)

	snapshot := dpdSignGoldenSnapshot{
		Headword:    entries[0].Headword,
		Markdown:    entries[0].Content,
		InlineKinds: kinds,
	}

	return json.MarshalIndent(snapshot, "", "  ")
}

func TestDPDSignsBracketSemanticTaggingUsesRealFixtures(t *testing.T) {
	tests := []struct {
		name          string
		term          string
		fixturePath   string
		wantKind      string
		wantBracketMD string
	}{
		{
			name:          "definition brackets from abrogar",
			term:          termAbrogar,
			fixturePath:   dpdFixturePath(termAbrogar),
			wantKind:      model.InlineKindBracketDefinition,
			wantBracketMD: "[una ley]",
		},
		{
			name:          "pronunciation brackets from alícuota",
			term:          termAlicuota,
			fixturePath:   dpdFixturePath(termAlicuota),
			wantKind:      model.InlineKindBracketPronunciation,
			wantBracketMD: "[alikuóto]",
		},
		{
			name:          "interpolation brackets from androfobia",
			term:          termAndrofobia,
			fixturePath:   dpdFixturePath(termAndrofobia),
			wantKind:      model.InlineKindBracketInterpolation,
			wantBracketMD: "[las feministas]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderDPDFixtureOutputs(t, tt.term, tt.fixturePath)
			requireJSONContainsInlineKind(t, output.json, tt.wantKind)
			requireMarkdownContainsPlainBracket(t, output.md, tt.wantBracketMD)
		})
	}
}

func TestDPDSignsBracketSemanticTaggingKeepsMixedContextsDistinctInOneArticle(t *testing.T) {
	raw := `<entry class="tem" id="mixed-brackets" header="mixed brackets"><header class="tem">mixed brackets</header><section class="BLOQUEACEPS"><p n="1n"><span class="enum">1.</span> Contextos: <dfn>[una ley]</dfn> <span class="nn">[alikuóto]</span> <span class="yy">[las feministas]</span>.</p></section></entry>`
	output := renderDPDRawHTMLOutputs(t, "mixed-brackets", raw)

	for _, kind := range []string{
		model.InlineKindBracketDefinition,
		model.InlineKindBracketPronunciation,
		model.InlineKindBracketInterpolation,
	} {
		requireJSONContainsInlineKind(t, output.json, kind)
	}

	for _, bracket := range []string{"[una ley]", "[alikuóto]", "[las feministas]"} {
		requireMarkdownContainsPlainBracket(t, output.md, bracket)
	}
}

func TestDPDSignsValidatedFixturesMatchGoldenOutputs(t *testing.T) {
	tests := []struct {
		name       string
		term       string
		fixture    string
		mdGolden   string
		jsonGolden string
	}{
		{
			name:       termAlicuota,
			term:       termAlicuota,
			fixture:    dpdFixturePath(termAlicuota),
			mdGolden:   dpdGoldenPath(termAlicuota, "md"),
			jsonGolden: dpdGoldenPath(termAlicuota, "json"),
		},
		{
			name:       termAcertar,
			term:       termAcertar,
			fixture:    dpdFixturePath(termAcertar),
			mdGolden:   dpdGoldenPath(termAcertar, "md"),
			jsonGolden: dpdGoldenPath(termAcertar, "json"),
		},
		{
			name:       termAbrogar,
			term:       termAbrogar,
			fixture:    dpdFixturePath(termAbrogar),
			mdGolden:   dpdGoldenPath(termAbrogar, "md"),
			jsonGolden: dpdGoldenPath(termAbrogar, "json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderDPDFixtureOutputs(t, tt.term, tt.fixture)
			requireMarkdownGolden(t, output.md, tt.mdGolden)
			snapshot, err := buildSignGoldenSnapshot(output.entries)
			if err != nil {
				t.Fatalf("buildSignGoldenSnapshot() error = %v", err)
			}
			requireJSONGolden(t, snapshot, tt.jsonGolden)
		})
	}
}

func TestDPDSignsValidatedFixturesDoNotRegressExclusionAndReferenceMarkers(t *testing.T) {
	alicuota := renderDPDFixtureOutputs(t, termAlicuota, dpdFixturePath(termAlicuota))
	if !strings.Contains(alicuota.md, exclusionMarker) {
		t.Fatalf("alícuota markdown lost exclusion marker\n%s", alicuota.md)
	}
	requireJSONContainsInlineKind(t, alicuota.json, model.InlineKindExclusion)

	acertar := renderDPDFixtureOutputs(t, termAcertar, dpdFixturePath(termAcertar))
	if !strings.Contains(acertar.md, acertarReference) {
		t.Fatalf("acertar markdown lost canonical arrow reference\n%s", acertar.md)
	}
	if !strings.Contains(acertar.md, exclusionMarker) {
		t.Fatalf("acertar markdown lost exclusion marker\n%s", acertar.md)
	}
	requireJSONContainsInlineKind(t, acertar.json, model.InlineKindReference)
	requireJSONContainsInlineKind(t, acertar.json, model.InlineKindExclusion)
}
