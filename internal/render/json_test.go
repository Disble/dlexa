package render

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

type jsonRendererArticleCitation struct {
	SourceLabel  string `json:"SourceLabel"`
	CanonicalURL string `json:"CanonicalURL"`
	Edition      string `json:"Edition"`
	ConsultedAt  string `json:"ConsultedAt"`
}

type jsonRendererArticlePayload struct {
	Entries []jsonRendererEntry `json:"Entries"`
}

type jsonRendererEntry struct {
	Content string              `json:"Content"`
	Article jsonRendererArticle `json:"Article"`
}

type jsonRendererArticle struct {
	Dictionary string                       `json:"Dictionary"`
	Edition    string                       `json:"Edition"`
	Lemma      string                       `json:"Lemma"`
	Sections   []jsonRendererArticleSection `json:"Sections"`
	Citation   jsonRendererArticleCitation  `json:"Citation"`
}

type jsonRendererArticleSection struct {
	Label    string                     `json:"Label"`
	Title    string                     `json:"Title"`
	Children []jsonRendererSectionChild `json:"Children"`
	Blocks   []jsonRendererSectionBlock `json:"Blocks"`
}

type jsonRendererSectionChild struct {
	Label string `json:"Label"`
}

type jsonRendererSectionBlock struct {
	Kind      string                 `json:"kind"`
	Paragraph *jsonRendererParagraph `json:"paragraph"`
}

type jsonRendererParagraph struct {
	Markdown string               `json:"Markdown"`
	Inlines  []jsonRendererInline `json:"Inlines"`
}

type jsonRendererInline struct {
	Kind   string `json:"Kind"`
	Text   string `json:"Text"`
	Target string `json:"Target"`
}

func TestJSONRendererSerializesArticleHierarchyAndCitation(t *testing.T) {
	renderer := NewJSONRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "bien", Format: "json"},
		Entries: []model.Entry{{ID: "dpd:bien#bien", Headword: "bien", Source: "dpd", URL: "https://www.rae.es/dpd/bien", Article: sampleBienArticle()}},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	decoded := decodeJSONRendererPayload(t, payload)
	entry := requireSingleJSONRendererEntry(t, decoded)
	assertJSONArticleIdentity(t, entry.Article)
	assertJSONArticleSections(t, entry.Article)
	assertJSONArticleReferenceInline(t, entry.Article)
	assertJSONArticleCitation(t, entry.Article.Citation)
	if entry.Content == "" {
		t.Fatal("content projection = empty, want article projection for compatibility")
	}
}

func TestDPDParseNormalizeRenderMatchesBienJSONGolden(t *testing.T) {
	entries := parseNormalizeDPD(t, "bien")
	jsonPayload, err := NewJSONRenderer().Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "bien", Format: "json"},
		Entries: entries,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	jsonWant, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", "bien.json.golden"))
	if err != nil {
		t.Fatalf("ReadFile() json golden error = %v\n--- got ---\n%s", err, jsonPayload)
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

func decodeJSONRendererPayload(t *testing.T, payload []byte) jsonRendererArticlePayload {
	t.Helper()

	var decoded jsonRendererArticlePayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return decoded
}

func requireSingleJSONRendererEntry(t *testing.T, decoded jsonRendererArticlePayload) jsonRendererEntry {
	t.Helper()

	if len(decoded.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(decoded.Entries))
	}
	return decoded.Entries[0]
}

func assertJSONArticleIdentity(t *testing.T, article jsonRendererArticle) {
	t.Helper()

	if article.Dictionary != testDictionary || article.Edition != testEdition || article.Lemma != "" {
		t.Fatalf("article identity = %#v", article)
	}
}

func assertJSONArticleSections(t *testing.T, article jsonRendererArticle) {
	t.Helper()

	if len(article.Sections) != 7 {
		t.Fatalf("sections = %d, want 7", len(article.Sections))
	}
	if article.Sections[4].Title != "bien que." || article.Sections[5].Title != "más bien." || article.Sections[6].Title != "si bien." {
		t.Fatalf("lexical titles = [%q %q %q]", article.Sections[4].Title, article.Sections[5].Title, article.Sections[6].Title)
	}
	if len(article.Sections[5].Children) != 3 {
		t.Fatalf("section 6 children = %d, want 3", len(article.Sections[5].Children))
	}
}

func assertJSONArticleReferenceInline(t *testing.T, article jsonRendererArticle) {
	t.Helper()

	firstParagraph := article.Sections[0].Blocks[0].Paragraph
	if firstParagraph == nil {
		t.Fatal("first paragraph = nil")
	}
	if len(firstParagraph.Inlines) == 0 {
		t.Fatal("first paragraph inlines = 0, want semantic inlines")
	}
	if !hasJSONReferenceInline(firstParagraph.Inlines) {
		t.Fatalf("first paragraph missing inline reference semantics: %#v", firstParagraph.Inlines)
	}
}

func hasJSONReferenceInline(inlines []jsonRendererInline) bool {
	for _, inline := range inlines {
		if inline.Kind == model.InlineKindReference && inline.Text == "6" && inline.Target == "bien#S1590507271213267522" {
			return true
		}
	}
	return false
}

func assertJSONArticleCitation(t *testing.T, citation jsonRendererArticleCitation) {
	t.Helper()

	if citation.SourceLabel != testSourceLabel || citation.CanonicalURL != "https://www.rae.es/dpd/bien" || citation.Edition != testEdition || citation.ConsultedAt != testConsultedAt {
		t.Fatalf("citation = %#v", citation)
	}
}
