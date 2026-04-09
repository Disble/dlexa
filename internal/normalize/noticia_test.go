package normalize

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

func TestNoticiaNormalizerBuildsLookupEntry(t *testing.T) {
	normalizer := NewNoticiaNormalizer()
	result, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: "noticia"}, parse.Result{Articles: []parse.ParsedArticle{{
		Dictionary:   "Noticia RAE",
		Lemma:        "Preguntas frecuentes: tilde en las mayúsculas",
		CanonicalURL: "https://www.rae.es/noticia/preguntas-frecuentes-tilde-en-las-mayusculas",
		Sections: []parse.ParsedSection{{
			Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: `<p>Respuesta breve con <strong>énfasis</strong>.</p>`}}},
		}},
	}}})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(result.Entries))
	}
	entry := result.Entries[0]
	if entry.Source != "noticia" {
		t.Fatalf("Source = %q, want noticia", entry.Source)
	}
	if entry.Article == nil {
		t.Fatal("Article = nil, want structured article")
	}
	if !strings.Contains(entry.Content, "Respuesta breve con énfasis.") {
		t.Fatalf("Content = %q, want normalized paragraph text", entry.Content)
	}
}

func TestNoticiaNormalizerRejectsArticlesWithoutSections(t *testing.T) {
	_, err := NewNoticiaNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "noticia"}, parse.Result{Articles: []parse.ParsedArticle{{Lemma: "faq"}}})
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
