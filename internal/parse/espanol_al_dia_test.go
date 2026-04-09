package parse

import (
	"context"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

func TestEspanolAlDiaParserExtractsArticleContent(t *testing.T) {
	parser := NewEspanolAlDiaParser()
	result, warnings, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "espanol-al-dia"}, fetch.Document{
		URL:  "https://www.rae.es/espanol-al-dia/el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde",
		Body: []byte(`<div class="container pt-8 pb-8 bloque-texto"><div class="row"><div class="col-md-8"><p class="antetitulo pb-2"><a href="/espanol-al-dia" class="link-migas">ESPAÑOL AL DÍA</a></p><h1 class="news-title"><span>El adverbio «solo» y los pronombres demostrativos, sin tilde</span></h1></div></div><div class="row"><div class="col-md-8"><div class="pt-4"><p>Primer párrafo con <em>énfasis</em>.</p><p>Segundo párrafo.</p></div></div></div></div><section class="section section--secondary">También le interesará</section>`),
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
	if article.Dictionary != "Español al día" {
		t.Fatalf("Dictionary = %q, want Español al día", article.Dictionary)
	}
	if article.Lemma != "El adverbio «solo» y los pronombres demostrativos, sin tilde" {
		t.Fatalf("Lemma = %q", article.Lemma)
	}
	if article.CanonicalURL != "https://www.rae.es/espanol-al-dia/el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde" {
		t.Fatalf("CanonicalURL = %q", article.CanonicalURL)
	}
	if len(article.Sections) != 1 || len(article.Sections[0].Blocks) != 2 {
		t.Fatalf("Sections = %#v, want one section with two paragraph blocks", article.Sections)
	}
}

func TestEspanolAlDiaParserReportsBrokenMarkupExplicitly(t *testing.T) {
	_, _, err := NewEspanolAlDiaParser().Parse(context.Background(), model.SourceDescriptor{Name: "espanol-al-dia"}, fetch.Document{
		URL:  "https://www.rae.es/espanol-al-dia/solo",
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
