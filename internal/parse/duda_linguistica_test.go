package parse

import (
	"context"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

func TestDudaLinguisticaParserExtractsArticleContent(t *testing.T) {
	parser := NewDudaLinguisticaParser()
	result, warnings, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "duda-linguistica"}, fetch.Document{
		URL:  "https://www.rae.es/duda-linguistica/cuando-se-escriben-con-tilde-los-adverbios-en-mente",
		Body: []byte(`<div class="container pt-8 pb-8 page-seccion bloque-texto"><div class="row"><div class="col-md-8"><p class="antetitulo pb-2"><a href="/portal-linguistico/dudas-rapidas" class="link-migas">Dudas rápidas</a></p><h1 class="news-title"><span>¿Cuándo se escriben con tilde los adverbios en «-mente»?</span></h1></div></div><div class="row"><div class="col-md-8"><div class="pt-4"><p>Los adverbios en <em>-mente</em> solo se escriben con tilde cuando la lleva el adjetivo base.</p></div></div></div></div><section class="section section--secondary">También le interesará</section>`),
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
	if article.Dictionary != "Duda lingüística" {
		t.Fatalf("Dictionary = %q, want Duda lingüística", article.Dictionary)
	}
	if article.Lemma != "¿Cuándo se escriben con tilde los adverbios en «-mente»?" {
		t.Fatalf("Lemma = %q", article.Lemma)
	}
	if article.CanonicalURL != "https://www.rae.es/duda-linguistica/cuando-se-escriben-con-tilde-los-adverbios-en-mente" {
		t.Fatalf("CanonicalURL = %q", article.CanonicalURL)
	}
	if len(article.Sections) != 1 || len(article.Sections[0].Blocks) != 1 {
		t.Fatalf("Sections = %#v, want one section with one paragraph block", article.Sections)
	}
}

func TestDudaLinguisticaParserReportsBrokenMarkupExplicitly(t *testing.T) {
	_, _, err := NewDudaLinguisticaParser().Parse(context.Background(), model.SourceDescriptor{Name: "duda-linguistica"}, fetch.Document{
		URL:  "https://www.rae.es/duda-linguistica/solo",
		Body: []byte(`<html><body><h1>sin estructura esperada</h1></body></html>`),
	})
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
