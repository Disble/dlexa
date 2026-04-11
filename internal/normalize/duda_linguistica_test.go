package normalize

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

const normalizeDudaLinguisticaSource = "duda-linguistica"

func TestDudaLinguisticaNormalizerBuildsLookupEntry(t *testing.T) {
	normalizer := NewDudaLinguisticaNormalizer()
	result, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: normalizeDudaLinguisticaSource}, parse.Result{Articles: []parse.ParsedArticle{{
		Dictionary:   "Duda lingüística",
		Lemma:        "¿Cuándo se escriben con tilde los adverbios en «-mente»?",
		CanonicalURL: "https://www.rae.es/duda-linguistica/cuando-se-escriben-con-tilde-los-adverbios-en-mente",
		Sections: []parse.ParsedSection{{
			Blocks: []parse.ParsedBlock{
				{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: `<p>Respuesta breve con <em>énfasis</em>.</p>`}},
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
	if entry.Source != normalizeDudaLinguisticaSource {
		t.Fatalf("Source = %q, want duda-linguistica", entry.Source)
	}
	if entry.Article == nil {
		t.Fatal("Article = nil, want structured article")
	}
	if !strings.Contains(entry.Content, "Respuesta breve con *énfasis*.") {
		t.Fatalf("Content = %q, want normalized paragraph text", entry.Content)
	}
}

func TestDudaLinguisticaNormalizerRejectsArticlesWithoutSections(t *testing.T) {
	_, err := NewDudaLinguisticaNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: normalizeDudaLinguisticaSource}, parse.Result{Articles: []parse.ParsedArticle{{Lemma: "tilde"}}})
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
