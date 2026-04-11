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
	"github.com/Disble/dlexa/internal/doctor"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	searchmodule "github.com/Disble/dlexa/internal/modules/search"
	"github.com/Disble/dlexa/internal/platform"
	"github.com/Disble/dlexa/internal/render"
	"github.com/Disble/dlexa/internal/version"
)

func TestAppExecuteModuleWrapsMarkdownAndBypassesJSON(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", DefaultLookupSources: []string{"dpd"}, Search: config.SearchConfig{DefaultProviders: []string{"search"}}, CacheEnabled: true}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	application := NewWithDependencies(
		cli,
		loader,
		&appDoctor{},
		modules.NewRegistry(
			&stubModule{command: "dpd", response: modules.Response{Title: "solo", Source: "Diccionario panhispánico de dudas", CacheState: "MISS", Format: "markdown", Body: []byte("## Resultado\ncontenido")}},
			&stubModule{command: commandEspanolAlDia, response: modules.Response{Title: "solo", Source: "Español al día", CacheState: "MISS", Format: "markdown", Body: []byte("## Artículo\ncontenido")}},
			&stubModule{command: commandDudaLinguistica, response: modules.Response{Title: "tilde", Source: "Duda lingüística", CacheState: "MISS", Format: "markdown", Body: []byte("## Duda\ncontenido")}},
			&stubModule{command: "noticia", response: modules.Response{Title: "faq", Source: "Preguntas frecuentes RAE", CacheState: "MISS", Format: "markdown", Body: []byte("## FAQ\ncontenido")}},
			&stubModule{command: "search", response: modules.Response{Title: "solo", Source: sourceBusquedaGeneralRAE, CacheState: "HIT", Format: "json", Body: []byte(`{"ok":true}`)}},
		),
		render.NewEnvelopeRenderer(),
	)

	if err := application.ExecuteModule(context.Background(), "dpd", modules.Request{Query: "solo"}); err != nil {
		t.Fatalf("ExecuteModule() markdown error = %v", err)
	}
	markdownText := cli.stdout.String()
	for _, want := range []string{"# [dlexa:dpd] solo", "*Fuente: Diccionario panhispánico de dudas | Caché: MISS*", "## Resultado"} {
		if !strings.Contains(markdownText, want) {
			t.Fatalf("markdown stdout missing %q\n%s", want, markdownText)
		}
	}

	cli.stdout.Reset()
	if err := application.ExecuteModule(context.Background(), "search", modules.Request{Query: "solo", Format: "json"}); err != nil {
		t.Fatalf("ExecuteModule() json error = %v", err)
	}
	if got := strings.TrimSpace(cli.stdout.String()); got != `{"ok":true}` {
		t.Fatalf("json stdout = %q, want untouched json body", got)
	}
}

func TestAppHandlesStructuredFallbacksAndSyntaxErrors(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", DefaultLookupSources: []string{"dpd"}, Search: config.SearchConfig{DefaultProviders: []string{"search"}}, CacheEnabled: true}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	application := NewWithDependencies(
		cli,
		loader,
		&appDoctor{},
		modules.NewRegistry(
			&stubModule{command: "dpd", response: modules.Response{Title: "solo", Source: "Diccionario panhispánico de dudas", CacheState: "MISS", Format: "markdown", Fallback: &model.FallbackEnvelope{Kind: model.FallbackKindNotFound, Module: "dpd", Query: "solo", Message: "No se encontró contenido en este módulo.", NextCommand: "dlexa search solo"}}},
		),
		render.NewEnvelopeRenderer(),
	)

	if err := application.ExecuteModule(context.Background(), "dpd", modules.Request{Query: "solo"}); err != nil {
		t.Fatalf("ExecuteModule() fallback error = %v", err)
	}
	if text := cli.stdout.String(); !strings.Contains(text, "Nivel 2 · Not Found") || !strings.Contains(text, "dlexa search solo") {
		t.Fatalf("fallback stdout = %q", text)
	}

	cli.stdout.Reset()
	if err := application.HandleSyntaxError(context.Background(), errors.New("unknown command \"oops\" for \"dlexa\""), helpSyntaxDlexaQuery); err != nil {
		t.Fatalf("HandleSyntaxError() error = %v", err)
	}
	if text := cli.stdout.String(); !strings.Contains(text, "Nivel 1 · Syntax") || !strings.Contains(text, helpSyntaxDlexaQuery) {
		t.Fatalf("syntax stdout = %q", text)
	}
}

func TestAppRendersMarkdownHelpAndDoctorOutput(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", DefaultLookupSources: []string{"dpd"}, Search: config.SearchConfig{DefaultProviders: []string{"search"}}, CacheEnabled: true}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	application := NewWithDependencies(
		cli,
		loader,
		&appDoctor{report: doctor.Report{Healthy: true, Checks: []doctor.Check{{Name: "bootstrap", Status: "ok", Detail: "doctor wiring is ready"}}}},
		modules.NewRegistry(),
		render.NewEnvelopeRenderer(),
	)

	if err := application.RenderHelp(context.Background(), model.HelpEnvelope{Command: "dlexa", Summary: "Consulta dudas normativas del español.", Syntax: helpSyntaxDlexaQuery, Examples: []string{"dlexa basto", "dlexa search solo o sólo"}, RecoveryTip: "Usá `dlexa search <consulta>` cuando no conozcas la entrada exacta."}); err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}
	if text := cli.stdout.String(); !strings.Contains(text, "# Ayuda: dlexa") || !strings.Contains(text, "`dlexa basto`") {
		t.Fatalf("help stdout = %q", text)
	}

	cli.stdout.Reset()
	if err := application.RunDoctor(context.Background()); err != nil {
		t.Fatalf("RunDoctor() error = %v", err)
	}
	if text := cli.stdout.String(); !strings.Contains(text, "doctor: ok") || !strings.Contains(text, "bootstrap [ok] doctor wiring is ready") {
		t.Fatalf("doctor stdout = %q", text)
	}
}

type fakeCLI struct {
	args   []string
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func (c *fakeCLI) Args() []string    { return c.args }
func (c *fakeCLI) Stdout() io.Writer { return &c.stdout }
func (c *fakeCLI) Stderr() io.Writer { return &c.stderr }

type appLoader struct{ cfg config.RuntimeConfig }

func (l *appLoader) Load(context.Context) (config.RuntimeConfig, error) { return l.cfg, nil }

type appDoctor struct{ report doctor.Report }

func (d *appDoctor) Run(context.Context) (doctor.Report, error) {
	if len(d.report.Checks) == 0 {
		d.report = doctor.Report{Healthy: true, Checks: []doctor.Check{{Name: "bootstrap", Status: "ok", Detail: "doctor wiring is ready"}}}
	}
	return d.report, nil
}

type stubModule struct {
	command     string
	response    modules.Response
	err         error
	lastRequest modules.Request
}

func (s stubModule) Name() string    { return s.command }
func (s stubModule) Command() string { return s.command }
func (s *stubModule) Execute(_ context.Context, req modules.Request) (modules.Response, error) {
	s.lastRequest = req
	return s.response, s.err
}

var _ platform.CLI = (*fakeCLI)(nil)

func TestExecuteModuleRejectsInvalidFormat(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", DefaultLookupSources: []string{"dpd"}, Search: config.SearchConfig{DefaultProviders: []string{"search"}}, CacheEnabled: true}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	application := NewWithDependencies(
		cli,
		loader,
		&appDoctor{},
		modules.NewRegistry(
			&stubModule{command: "dpd", response: modules.Response{Title: "solo", Source: "DPD", CacheState: "MISS", Format: "markdown", Body: []byte("contenido")}},
		),
		render.NewEnvelopeRenderer(),
	)

	// yaml is not a registered format — should produce a Nivel 1 Syntax fallback, not a raw error.
	if err := application.ExecuteModule(context.Background(), "dpd", modules.Request{Query: "solo", Format: "yaml"}); err != nil {
		t.Fatalf("ExecuteModule() should not return error for invalid format, got %v", err)
	}
	text := cli.stdout.String()
	if !strings.Contains(text, "Nivel 1 · Syntax") {
		t.Fatalf("expected Nivel 1 Syntax fallback for invalid format, got:\n%s", text)
	}
	if !strings.Contains(text, "markdown") || !strings.Contains(text, "json") {
		t.Fatalf("expected fallback to mention supported formats, got:\n%s", text)
	}
}

func TestExecuteModuleAppliesModuleSpecificDefaultSources(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{
		DefaultFormat:        "markdown",
		DefaultLookupSources: []string{"dpd"},
		Search: config.SearchConfig{
			DefaultProviders: []string{"search"},
		},
		CacheEnabled: true,
	}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	dpdModule := &stubModule{command: "dpd", response: modules.Response{Title: "basto", Source: "DPD", CacheState: "MISS", Format: "markdown", Body: []byte("contenido")}}
	searchModule := &stubModule{command: "search", response: modules.Response{Title: "basto", Source: sourceBusquedaGeneralRAE, CacheState: "MISS", Format: "json", Body: []byte(`{"ok":true}`)}}
	application := NewWithDependencies(cli, loader, &appDoctor{}, modules.NewRegistry(dpdModule, searchModule), render.NewEnvelopeRenderer())

	if err := application.ExecuteModule(context.Background(), "dpd", modules.Request{Query: "basto"}); err != nil {
		t.Fatalf("ExecuteModule() dpd error = %v", err)
	}
	if got := dpdModule.lastRequest.Sources; !reflect.DeepEqual(got, []string{"dpd"}) {
		t.Fatalf("dpd sources = %#v, want [\"dpd\"]", got)
	}

	if err := application.ExecuteModule(context.Background(), "search", modules.Request{Query: "basto", Format: "json"}); err != nil {
		t.Fatalf("ExecuteModule() search error = %v", err)
	}
	if got := searchModule.lastRequest.Sources; !reflect.DeepEqual(got, []string{"search"}) {
		t.Fatalf("search sources = %#v, want [\"search\"]", got)
	}

	explicit := []string{"manual"}
	if err := application.ExecuteModule(context.Background(), "search", modules.Request{Query: "basto", Format: "json", Sources: explicit}); err != nil {
		t.Fatalf("ExecuteModule() explicit sources error = %v", err)
	}
	if got := searchModule.lastRequest.Sources; !reflect.DeepEqual(got, explicit) {
		t.Fatalf("explicit search sources = %#v, want %#v", got, explicit)
	}
}

func TestExecuteModuleAppliesFederatedSearchDefaults(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{
		DefaultFormat:        "markdown",
		DefaultLookupSources: []string{"dpd"},
		Search: config.SearchConfig{
			DefaultProviders: []string{"search", "dpd"},
		},
		CacheEnabled: true,
	}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	searchModule := &stubModule{command: "search", response: modules.Response{Title: "Abu Dhabi", Source: sourceBusquedaGeneralRAE, CacheState: "MISS", Format: "json", Body: []byte(`{"ok":true}`)}}
	application := NewWithDependencies(cli, loader, &appDoctor{}, modules.NewRegistry(searchModule), render.NewEnvelopeRenderer())

	if err := application.ExecuteModule(context.Background(), "search", modules.Request{Query: "Abu Dhabi", Format: "json"}); err != nil {
		t.Fatalf("ExecuteModule() search error = %v", err)
	}
	if got := searchModule.lastRequest.Sources; !reflect.DeepEqual(got, []string{"search", "dpd"}) {
		t.Fatalf("search sources = %#v, want federated defaults [\"search\", \"dpd\"]", got)
	}
}

func TestExecuteModuleUsesLookupDefaultsForEspanolAlDia(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{
		DefaultFormat:        "markdown",
		DefaultLookupSources: []string{"dpd"},
		Search: config.SearchConfig{
			DefaultProviders: []string{"search", "dpd"},
		},
		CacheEnabled: true,
	}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	eadModule := &stubModule{command: commandEspanolAlDia, response: modules.Response{Title: "solo", Source: "Español al día", CacheState: "MISS", Format: "markdown", Body: []byte("contenido")}}
	application := NewWithDependencies(cli, loader, &appDoctor{}, modules.NewRegistry(eadModule), render.NewEnvelopeRenderer())

	if err := application.ExecuteModule(context.Background(), commandEspanolAlDia, modules.Request{Query: "solo"}); err != nil {
		t.Fatalf(errExecuteModule, err)
	}
	if got := eadModule.lastRequest.Sources; !reflect.DeepEqual(got, []string{commandEspanolAlDia}) {
		t.Fatalf("sources = %#v, want module-specific default [\"%s\"]", got, commandEspanolAlDia)
	}
}

func TestExecuteModuleUsesLookupDefaultsForDudaLinguistica(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{
		DefaultFormat:        "markdown",
		DefaultLookupSources: []string{"dpd"},
		Search: config.SearchConfig{
			DefaultProviders: []string{"search", "dpd"},
		},
		CacheEnabled: true,
	}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	dlModule := &stubModule{command: commandDudaLinguistica, response: modules.Response{Title: "tilde", Source: "Duda lingüística", CacheState: "MISS", Format: "markdown", Body: []byte("contenido")}}
	application := NewWithDependencies(cli, loader, &appDoctor{}, modules.NewRegistry(dlModule), render.NewEnvelopeRenderer())

	if err := application.ExecuteModule(context.Background(), commandDudaLinguistica, modules.Request{Query: "tilde"}); err != nil {
		t.Fatalf(errExecuteModule, err)
	}
	if got := dlModule.lastRequest.Sources; !reflect.DeepEqual(got, []string{commandDudaLinguistica}) {
		t.Fatalf("sources = %#v, want module-specific default [\"%s\"]", got, commandDudaLinguistica)
	}
}

func TestExecuteModuleUsesLookupDefaultsForNoticia(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{
		DefaultFormat:        "markdown",
		DefaultLookupSources: []string{"dpd"},
		Search: config.SearchConfig{
			DefaultProviders: []string{"search", "dpd"},
		},
		CacheEnabled: true,
	}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	noticiaModule := &stubModule{command: "noticia", response: modules.Response{Title: "faq", Source: "Preguntas frecuentes RAE", CacheState: "MISS", Format: "markdown", Body: []byte("contenido")}}
	application := NewWithDependencies(cli, loader, &appDoctor{}, modules.NewRegistry(noticiaModule), render.NewEnvelopeRenderer())

	if err := application.ExecuteModule(context.Background(), "noticia", modules.Request{Query: "preguntas-frecuentes-tilde-en-las-mayusculas"}); err != nil {
		t.Fatalf(errExecuteModule, err)
	}
	if got := noticiaModule.lastRequest.Sources; !reflect.DeepEqual(got, []string{"noticia"}) {
		t.Fatalf("sources = %#v, want module-specific default [\"noticia\"]", got)
	}
}

func TestExecuteModuleWrapsDPDSearchWithTruthfulSource(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{
		DefaultFormat:        "markdown",
		DefaultLookupSources: []string{"dpd"},
		Search: config.SearchConfig{
			DefaultProviders: []string{"search", "dpd"},
		},
		CacheEnabled: true,
	}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	searcher := &appSearcherStub{result: model.SearchResult{Request: model.SearchRequest{Query: "solo", Format: "markdown", Sources: []string{"dpd"}}}}
	renderer := &appSearchRendererStub{payload: []byte("## Resultado semántico\ncontenido")}
	searchModule := searchmodule.New(searcher, &appSearchRenderersStub{renderer: renderer})
	application := NewWithDependencies(cli, loader, &appDoctor{}, modules.NewRegistry(searchModule), render.NewEnvelopeRenderer())

	if err := application.ExecuteModule(context.Background(), "search", modules.Request{Query: "solo", Sources: []string{"dpd"}}); err != nil {
		t.Fatalf(errExecuteModule, err)
	}
	text := cli.stdout.String()
	if !strings.Contains(text, "*Fuente: Diccionario panhispánico de dudas | Caché: MISS*") {
		t.Fatalf("expected truthful DPD source in markdown envelope, got:\n%s", text)
	}
	if strings.Contains(text, "*Fuente: "+sourceBusquedaGeneralRAE+" | Caché: MISS*") {
		t.Fatalf("expected not to use general search source for DPD-only search, got:\n%s", text)
	}
}

type appSearcherStub struct {
	result model.SearchResult
	err    error
}

func (s *appSearcherStub) Search(context.Context, model.SearchRequest) (model.SearchResult, error) {
	return s.result, s.err
}

type appSearchRenderersStub struct{ renderer render.SearchRenderer }

func (s *appSearchRenderersStub) Renderer(string) (render.SearchRenderer, error) {
	return s.renderer, nil
}

type appSearchRendererStub struct{ payload []byte }

func (s *appSearchRendererStub) Format() string { return "markdown" }

func (s *appSearchRendererStub) Render(context.Context, model.SearchResult) ([]byte, error) {
	return s.payload, nil
}
