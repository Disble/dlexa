package parse

import (
	"context"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

func TestNoticiaParserExtractsArticleContent(t *testing.T) {
	parser := NewNoticiaParser()
	result, warnings, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "noticia"}, fetch.Document{
		URL:  "https://www.rae.es/noticia/preguntas-frecuentes-tilde-en-las-mayusculas",
		Body: []byte(`<div class="container-news"><h1 class="news-title">Preguntas frecuentes: tilde en las mayúsculas</h1><div class="bloque-texto"><p>Primer párrafo con <strong>norma</strong>.</p><p>Segundo párrafo.</p><p>&nbsp;</p></div></div>`),
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if len(result.Articles) != 1 {
		t.Fatalf("articles len = %d, want 1", len(result.Articles))
	}
	article := result.Articles[0]
	if article.Dictionary != "Noticia RAE" {
		t.Fatalf("Dictionary = %q, want Noticia RAE", article.Dictionary)
	}
	if article.Lemma != "Preguntas frecuentes: tilde en las mayúsculas" {
		t.Fatalf("Lemma = %q", article.Lemma)
	}
	if len(article.Sections) != 1 || len(article.Sections[0].Blocks) != 2 {
		t.Fatalf("Sections = %#v, want one section with two paragraph blocks", article.Sections)
	}
}

func TestNoticiaParserReportsBrokenMarkupExplicitly(t *testing.T) {
	_, _, err := NewNoticiaParser().Parse(context.Background(), model.SourceDescriptor{Name: "noticia"}, fetch.Document{URL: "https://www.rae.es/noticia/faq", Body: []byte(`<html><body><h1>sin estructura esperada</h1></body></html>`)})
	if err == nil {
		t.Fatal("Parse() error = nil, want problem")
	}
	problem, ok := model.AsProblem(err)
	if !ok {
		t.Fatalf("Parse() error = %T, want problem", err)
	}
	if problem.Code != model.ProblemCodeArticleExtractFailed {
		t.Fatalf("problem code = %q, want %q", problem.Code, model.ProblemCodeArticleExtractFailed)
	}
}
