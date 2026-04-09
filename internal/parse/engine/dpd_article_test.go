package engine

import (
	"context"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

func TestDPDArticleParser_UnderlyingForTesting(t *testing.T) {
	parser := NewDPDArticleParser()
	if parser.UnderlyingForTesting() == nil {
		t.Error("expected non-nil legacy parser")
	}
}

func TestDPDArticleParser_DelegatesParseArticle(t *testing.T) {
	parser := NewDPDArticleParser()
	input := ParseInput{
		Ctx:        context.Background(),
		Descriptor: model.SourceDescriptor{Name: "dpd"},
		Document: fetch.Document{Body: []byte(`
			<html>
			  <body>
			    <entry class="lex" id="solo"><header class="lex">solo</header><p><span class="enum">1.</span>Artículo de prueba.</p></entry>
			    <p class="o">Diccionario panhispánico de dudas</p>
			  </body>
			</html>`), URL: "https://www.rae.es/dpd/solo"},
	}

	gotResult, gotWarnings, err := parser.ParseArticle(input)
	if err != nil {
		t.Fatalf("ParseArticle() error = %v", err)
	}
	wantResult, wantWarnings, err := parser.UnderlyingForTesting().Parse(input.Ctx, input.Descriptor, input.Document)
	if err != nil {
		t.Fatalf("legacy Parse() error = %v", err)
	}
	if !reflect.DeepEqual(gotResult, wantResult) {
		t.Fatalf("ParseArticle() result = %#v, want %#v", gotResult, wantResult)
	}
	if !reflect.DeepEqual(gotWarnings, wantWarnings) {
		t.Fatalf("ParseArticle() warnings = %#v, want %#v", gotWarnings, wantWarnings)
	}
}
