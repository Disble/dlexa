package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/doctor"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/platform"
	"github.com/Disble/dlexa/internal/render"
	"github.com/Disble/dlexa/internal/version"
)

func TestAppExecuteModuleWrapsMarkdownAndBypassesJSON(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", DefaultSources: []string{"dpd"}, CacheEnabled: true}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	application := NewWithDependencies(
		cli,
		loader,
		&appDoctor{},
		modules.NewRegistry(
			stubModule{command: "dpd", response: modules.Response{Title: "solo", Source: "Diccionario panhispánico de dudas", CacheState: "MISS", Format: "markdown", Body: []byte("## Resultado\ncontenido")}},
			stubModule{command: "search", response: modules.Response{Title: "solo", Source: "búsqueda general RAE", CacheState: "HIT", Format: "json", Body: []byte(`{"ok":true}`)}},
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
	loader := &appLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", DefaultSources: []string{"dpd"}, CacheEnabled: true}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	application := NewWithDependencies(
		cli,
		loader,
		&appDoctor{},
		modules.NewRegistry(
			stubModule{command: "dpd", response: modules.Response{Title: "solo", Source: "Diccionario panhispánico de dudas", CacheState: "MISS", Format: "markdown", Fallback: &model.FallbackEnvelope{Kind: model.FallbackKindNotFound, Module: "dpd", Query: "solo", Message: "No se encontró contenido en este módulo.", NextCommand: "dlexa search solo"}}},
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
	if err := application.HandleSyntaxError(context.Background(), errors.New("unknown command \"oops\" for \"dlexa\""), "dlexa <query>"); err != nil {
		t.Fatalf("HandleSyntaxError() error = %v", err)
	}
	if text := cli.stdout.String(); !strings.Contains(text, "Nivel 1 · Syntax") || !strings.Contains(text, "dlexa <query>") {
		t.Fatalf("syntax stdout = %q", text)
	}
}

func TestAppRendersMarkdownHelpAndDoctorOutput(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", DefaultSources: []string{"dpd"}, CacheEnabled: true}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	application := NewWithDependencies(
		cli,
		loader,
		&appDoctor{report: doctor.Report{Healthy: true, Checks: []doctor.Check{{Name: "bootstrap", Status: "ok", Detail: "doctor wiring is ready"}}}},
		modules.NewRegistry(),
		render.NewEnvelopeRenderer(),
	)

	if err := application.RenderHelp(context.Background(), model.HelpEnvelope{Command: "dlexa", Summary: "Consulta dudas normativas del español.", Syntax: "dlexa <query>", Examples: []string{"dlexa basto", "dlexa search solo o sólo"}, RecoveryTip: "Usá `dlexa search <consulta>` cuando no conozcas la entrada exacta."}); err != nil {
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
	command  string
	response modules.Response
	err      error
}

func (s stubModule) Name() string    { return s.command }
func (s stubModule) Command() string { return s.command }
func (s stubModule) Execute(context.Context, modules.Request) (modules.Response, error) {
	return s.response, s.err
}

var _ platform.CLI = (*fakeCLI)(nil)

func TestExecuteModuleRejectsInvalidFormat(t *testing.T) {
	loader := &appLoader{cfg: config.RuntimeConfig{DefaultFormat: "markdown", DefaultSources: []string{"dpd"}, CacheEnabled: true}}
	cli := &fakeCLI{args: []string{version.BinaryName}}
	application := NewWithDependencies(
		cli,
		loader,
		&appDoctor{},
		modules.NewRegistry(
			stubModule{command: "dpd", response: modules.Response{Title: "solo", Source: "DPD", CacheState: "MISS", Format: "markdown", Body: []byte("contenido")}},
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
