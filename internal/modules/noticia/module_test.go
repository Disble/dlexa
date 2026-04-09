package noticia

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/render"
)

func TestModuleTargetsNoticiaSourceAndReturnsBody(t *testing.T) {
	lookup := &lookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "preguntas-frecuentes-tilde-en-las-mayusculas", Format: "markdown", Sources: []string{"noticia"}}, CacheHit: true, Entries: []model.Entry{{Headword: "Preguntas frecuentes: tilde en las mayúsculas", Content: "contenido", Article: &model.Article{Lemma: "Preguntas frecuentes: tilde en las mayúsculas"}}}}}
	renderers := &renderersStub{renderer: &rendererStub{payload: []byte("## Noticia\ncontenido")}}
	module := New(lookup, renderers)

	response, err := module.Execute(context.Background(), modules.Request{Query: "preguntas-frecuentes-tilde-en-las-mayusculas", Format: "markdown"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := lookup.request.Sources; len(got) != 1 || got[0] != "noticia" {
		t.Fatalf("lookup sources = %#v, want [\"noticia\"]", got)
	}
	if response.Source != "Preguntas frecuentes RAE" || string(response.Body) != "## Noticia\ncontenido" {
		t.Fatalf("response = %#v", response)
	}
}

func TestModuleRejectsNonFAQNoticiaContent(t *testing.T) {
	lookup := &lookupStub{result: model.LookupResult{Request: model.LookupRequest{Query: "geopolitica", Format: "markdown", Sources: []string{"noticia"}}, Entries: []model.Entry{{Headword: "Geopolítica del español", Content: "contenido", Article: &model.Article{Lemma: "La obra Geopolítica del español se presenta en la RAE"}}}}}
	module := New(lookup, &renderersStub{renderer: &rendererStub{payload: []byte("unused")}})

	response, err := module.Execute(context.Background(), modules.Request{Query: "geopolitica", Format: "markdown", Sources: []string{"noticia"}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.Fallback == nil || response.Fallback.Kind != model.FallbackKindNotFound {
		t.Fatalf("fallback = %#v, want not_found fallback", response.Fallback)
	}
	if !strings.Contains(response.Fallback.Message, "pregunta frecuente normativa") {
		t.Fatalf("fallback message = %q", response.Fallback.Message)
	}
}

type lookupStub struct {
	request model.LookupRequest
	result  model.LookupResult
	err     error
}

func (s *lookupStub) Lookup(_ context.Context, request model.LookupRequest) (model.LookupResult, error) {
	s.request = request
	return s.result, s.err
}

type renderersStub struct{ renderer render.Renderer }

func (s *renderersStub) Renderer(string) (render.Renderer, error) { return s.renderer, nil }

type rendererStub struct {
	payload []byte
	result  model.LookupResult
}

func (s *rendererStub) Format() string { return "markdown" }
func (s *rendererStub) Render(_ context.Context, result model.LookupResult) ([]byte, error) {
	s.result = result
	return s.payload, nil
}
