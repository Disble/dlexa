package app

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/gentleman-programming/dlexa/internal/config"
	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/render"
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
}

func (s *capturingLookupService) Lookup(_ context.Context, request model.LookupRequest) (model.LookupResult, error) {
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
