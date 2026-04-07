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

func TestJSONRendererSerializesArticleHierarchyAndCitation(t *testing.T) {
	renderer := NewJSONRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "bien", Format: "json"},
		Entries: []model.Entry{{ID: "dpd:bien#bien", Headword: "bien", Source: "dpd", URL: "https://www.rae.es/dpd/bien", Article: sampleBienArticle()}},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	var decoded struct {
		Entries []struct {
			Content string `json:"Content"`
			Article struct {
				Dictionary string `json:"Dictionary"`
				Edition    string `json:"Edition"`
				Lemma      string `json:"Lemma"`
				Sections   []struct {
					Label    string `json:"Label"`
					Title    string `json:"Title"`
					Children []struct {
						Label string `json:"Label"`
					} `json:"Children"`
					Blocks []struct {
						Kind      string `json:"kind"`
						Paragraph *struct {
							Markdown string `json:"Markdown"`
							Inlines  []struct {
								Kind   string `json:"Kind"`
								Text   string `json:"Text"`
								Target string `json:"Target"`
							} `json:"Inlines"`
						} `json:"paragraph"`
					} `json:"Blocks"`
				} `json:"Sections"`
				Citation struct {
					SourceLabel  string `json:"SourceLabel"`
					CanonicalURL string `json:"CanonicalURL"`
					Edition      string `json:"Edition"`
					ConsultedAt  string `json:"ConsultedAt"`
				} `json:"Citation"`
			} `json:"Article"`
		} `json:"Entries"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(decoded.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(decoded.Entries))
	}
	entry := decoded.Entries[0]
	if entry.Article.Dictionary != testDictionary || entry.Article.Edition != testEdition || entry.Article.Lemma != "" {
		t.Fatalf("article identity = %#v", entry.Article)
	}
	if len(entry.Article.Sections) != 7 {
		t.Fatalf("sections = %d, want 7", len(entry.Article.Sections))
	}
	if entry.Article.Sections[4].Title != "bien que." || entry.Article.Sections[5].Title != "más bien." || entry.Article.Sections[6].Title != "si bien." {
		t.Fatalf("lexical titles = [%q %q %q]", entry.Article.Sections[4].Title, entry.Article.Sections[5].Title, entry.Article.Sections[6].Title)
	}
	if len(entry.Article.Sections[5].Children) != 3 {
		t.Fatalf("section 6 children = %d, want 3", len(entry.Article.Sections[5].Children))
	}
	firstParagraph := entry.Article.Sections[0].Blocks[0].Paragraph
	if firstParagraph == nil {
		t.Fatal("first paragraph = nil")
	}
	if len(firstParagraph.Inlines) == 0 {
		t.Fatal("first paragraph inlines = 0, want semantic inlines")
	}
	foundReference := false
	for _, inline := range firstParagraph.Inlines {
		if inline.Kind == model.InlineKindReference && inline.Text == "6" && inline.Target == "bien#S1590507271213267522" {
			foundReference = true
			break
		}
	}
	if !foundReference {
		t.Fatalf("first paragraph missing inline reference semantics: %#v", firstParagraph.Inlines)
	}
	if entry.Article.Citation.SourceLabel != testSourceLabel || entry.Article.Citation.CanonicalURL != "https://www.rae.es/dpd/bien" || entry.Article.Citation.Edition != testEdition || entry.Article.Citation.ConsultedAt != testConsultedAt {
		t.Fatalf("citation = %#v", entry.Article.Citation)
	}
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
