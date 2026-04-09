package normalize

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

func TestEspanolAlDiaNormalizerBuildsLookupEntry(t *testing.T) {
	normalizer := NewEspanolAlDiaNormalizer()
	result, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: "espanol-al-dia"}, parse.Result{Articles: []parse.ParsedArticle{{
		Dictionary:   "Español al día",
		Lemma:        "El adverbio «solo» y los pronombres demostrativos, sin tilde",
		CanonicalURL: "https://www.rae.es/espanol-al-dia/el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde",
		Sections: []parse.ParsedSection{{
			Blocks: []parse.ParsedBlock{
				{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: `<p>Primer párrafo con <em>énfasis</em>.</p>`}},
				{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: `<p>Segundo párrafo.</p>`}},
			},
		}},
	}}})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(result.Entries))
	}
	entry := result.Entries[0]
	if entry.Source != "espanol-al-dia" {
		t.Fatalf("Source = %q, want espanol-al-dia", entry.Source)
	}
	if entry.Article == nil {
		t.Fatal("Article = nil, want structured article")
	}
	if !strings.Contains(entry.Content, "Primer párrafo con *énfasis*.") {
		t.Fatalf("Content = %q, want normalized paragraph text", entry.Content)
	}
}

func TestEspanolAlDiaNormalizerRejectsArticlesWithoutSections(t *testing.T) {
	_, err := NewEspanolAlDiaNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "espanol-al-dia"}, parse.Result{Articles: []parse.ParsedArticle{{Lemma: "solo"}}})
	if err == nil {
		t.Fatal("Normalize() error = nil, want problem")
	}
	problem, ok := model.AsProblem(err)
	if !ok {
		t.Fatalf("Normalize() error = %T, want problem", err)
	}
	if problem.Code != model.ProblemCodeArticleTransformFailed {
		t.Fatalf("problem code = %q, want %q", problem.Code, model.ProblemCodeArticleTransformFailed)
	}
}
