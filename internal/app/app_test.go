package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/render"
)

func TestRunConstructsLookupRequestAndDelegatesToQueryService(t *testing.T) {
	cli := &fakeCLI{
		args: []string{"dlexa", "--format", "json", "--source", "demo, extra ", "--no-cache", "hola", "mundo"},
	}
	lookup := &capturingLookupService{}
	renderer := &capturingRenderer{format: "json", payload: []byte("rendered output")}
	renderers := &capturingRegistry{renderer: renderer}

	application := &App{
		platform: cli,
		config: &staticLoader{cfg: config.RuntimeConfig{
			DefaultFormat:  "markdown",
			DefaultSources: []string{"fallback"},
			CacheEnabled:   true,
		}},
		lookup:    lookup,
		renderers: renderers,
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	wantRequest := model.LookupRequest{
		Query:   "hola mundo",
		Format:  "json",
		Sources: []string{"demo", "extra"},
		NoCache: true,
	}

	if !reflect.DeepEqual(lookup.request, wantRequest) {
		t.Fatalf("Lookup() request = %#v, want %#v", lookup.request, wantRequest)
	}

	if renderers.requestedFormat != "json" {
		t.Fatalf("Renderer() format = %q, want %q", renderers.requestedFormat, "json")
	}

	if !reflect.DeepEqual(renderer.result.Request, wantRequest) {
		t.Fatalf("Render() request = %#v, want %#v", renderer.result.Request, wantRequest)
	}

	if got := cli.stdout.String(); got != "rendered output\n" {
		t.Fatalf("stdout = %q, want %q", got, "rendered output\n")
	}
}

func TestRunDispatchesDedicatedSearchCommandWithoutLookupFlow(t *testing.T) {
	cli := &fakeCLI{args: []string{"dlexa", "--format", "json", "search", "abu", "dhabi"}}
	lookup := &capturingLookupService{}
	search := &capturingSearchService{result: model.SearchResult{Candidates: []model.SearchCandidate{{DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi"}}}}
	searchRenderer := &capturingSearchRenderer{format: "json", payload: []byte(`{"ok":true}`)}

	application := &App{
		platform:        cli,
		config:          &staticLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", CacheEnabled: true}},
		lookup:          lookup,
		search:          search,
		renderers:       &capturingRegistry{renderer: &capturingRenderer{format: "json"}},
		searchRenderers: &capturingSearchRegistry{renderer: searchRenderer},
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if lookup.called {
		t.Fatal("Lookup() was called for search command")
	}
	if !reflect.DeepEqual(search.request, model.SearchRequest{Query: "abu dhabi", Format: "json", NoCache: false}) {
		t.Fatalf("Search() request = %#v", search.request)
	}
	if got := cli.stdout.String(); got != "{\"ok\":true}\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunRejectsMissingSearchQueryLocally(t *testing.T) {
	cli := &fakeCLI{args: []string{"dlexa", "search"}}
	search := &capturingSearchService{}
	application := &App{
		platform:        cli,
		config:          &staticLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", CacheEnabled: true}},
		lookup:          &capturingLookupService{},
		search:          search,
		renderers:       &capturingRegistry{renderer: &capturingRenderer{format: "markdown"}},
		searchRenderers: &capturingSearchRegistry{renderer: &capturingSearchRenderer{format: "markdown"}},
	}

	err := application.Run(context.Background())
	if err == nil || err.Error() != "search command requires a query" {
		t.Fatalf("Run() error = %v, want missing-query error", err)
	}
	if search.called {
		t.Fatal("Search() was called for missing query")
	}
	if !strings.Contains(cli.stderr.String(), "search <query>") {
		t.Fatalf("stderr = %q", cli.stderr.String())
	}
}

func TestRunWritesRendererProducedStdoutPayloadForDPDSemantics(t *testing.T) {
	cli := &fakeCLI{args: []string{"dlexa", "bien"}}
	lookup := &capturingLookupService{}
	renderer := &capturingRenderer{
		format:  "markdown",
		payload: []byte("1. El comparativo es *mejor*. *Cierra bien la ventana*\n"),
	}
	renderers := &capturingRegistry{renderer: renderer}

	application := &App{
		platform: cli,
		config: &staticLoader{cfg: config.RuntimeConfig{
			DefaultFormat:  "markdown",
			DefaultSources: []string{"dpd"},
			CacheEnabled:   true,
		}},
		lookup:    lookup,
		renderers: renderers,
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	stdout := cli.stdout.String()
	if strings.Contains(stdout, "\x1b[") {
		t.Fatalf("stdout = %q, final boundary must not add ANSI noise by default", stdout)
	}
	if !strings.Contains(stdout, "El comparativo es *mejor*.") {
		t.Fatalf("stdout = %q, want renderer-produced markdown emphasis at final boundary", stdout)
	}
	if !strings.Contains(stdout, "*Cierra bien la ventana*") {
		t.Fatalf("stdout = %q, want renderer-produced markdown example at final boundary", stdout)
	}
	for _, forbidden := range []string{"[ej.:", "ej.:", "‹", "›"} {
		if strings.Contains(stdout, forbidden) {
			t.Fatalf("stdout = %q, contains forbidden fallback marker %q", stdout, forbidden)
		}
	}
	if got := renderer.result.Request.Query; got != "bien" {
		t.Fatalf("renderer query = %q, want %q", got, "bien")
	}
}

func TestRunPrintsSearchFirstUsageGuidanceWithoutImplyingFallback(t *testing.T) {
	cli := &fakeCLI{args: []string{"dlexa"}}
	application := &App{platform: cli}

	err := application.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	usage := cli.stderr.String()
	for _, want := range []string{
		"Use `dlexa search <query>` when you do not know the exact DPD entry yet.",
		"Use `dlexa <query>` when you already know the exact entry.",
	} {
		if !strings.Contains(usage, want) {
			t.Fatalf("usage missing %q\n%s", want, usage)
		}
	}
	for _, forbidden := range []string{"automatically falls back", "reroutes to search"} {
		if strings.Contains(usage, forbidden) {
			t.Fatalf("usage contains forbidden text %q\n%s", forbidden, usage)
		}
	}
}

type fakeCLI struct {
	args   []string
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func (c *fakeCLI) Args() []string {
	return c.args
}

func (c *fakeCLI) Stdout() io.Writer {
	return &c.stdout
}

func (c *fakeCLI) Stderr() io.Writer {
	return &c.stderr
}

type staticLoader struct {
	cfg config.RuntimeConfig
}

func (l *staticLoader) Load(context.Context) (config.RuntimeConfig, error) {
	return l.cfg, nil
}

type capturingLookupService struct {
	request model.LookupRequest
	called  bool
}

func (s *capturingLookupService) Lookup(_ context.Context, request model.LookupRequest) (model.LookupResult, error) {
	s.called = true
	s.request = request
	return model.LookupResult{Request: request}, nil
}

type capturingRegistry struct {
	renderer        *capturingRenderer
	requestedFormat string
}

func (r *capturingRegistry) Renderer(format string) (render.Renderer, error) {
	r.requestedFormat = format
	return r.renderer, nil
}

type capturingRenderer struct {
	format  string
	payload []byte
	result  model.LookupResult
}

func (r *capturingRenderer) Format() string {
	return r.format
}

func (r *capturingRenderer) Render(_ context.Context, result model.LookupResult) ([]byte, error) {
	r.result = result
	return r.payload, nil
}

type capturingSearchService struct {
	request model.SearchRequest
	result  model.SearchResult
	err     error
	called  bool
}

func (s *capturingSearchService) Search(_ context.Context, request model.SearchRequest) (model.SearchResult, error) {
	s.called = true
	s.request = request
	if s.err != nil {
		return model.SearchResult{}, s.err
	}
	if s.result.Request.Query == "" {
		s.result.Request = request
	}
	return s.result, nil
}

type capturingSearchRegistry struct {
	renderer        *capturingSearchRenderer
	requestedFormat string
}

func (r *capturingSearchRegistry) Renderer(format string) (render.SearchRenderer, error) {
	if r.renderer == nil {
		return nil, errors.New("missing search renderer")
	}
	r.requestedFormat = format
	return r.renderer, nil
}

type capturingSearchRenderer struct {
	format  string
	payload []byte
	result  model.SearchResult
}

func (r *capturingSearchRenderer) Format() string {
	return r.format
}

func (r *capturingSearchRenderer) Render(_ context.Context, result model.SearchResult) ([]byte, error) {
	r.result = result
	return r.payload, nil
}
