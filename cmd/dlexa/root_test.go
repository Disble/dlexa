package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/render"
)

// stubRuntime implements runtimeRunner for black-box command surface tests.
// It records which methods were called and, when stdout is set, writes
// real rendered output through a MarkdownEnvelopeRenderer so assertions
// can verify actual CLI output.
type stubRuntime struct {
	// stdout receives rendered output when non-nil.
	stdout *bytes.Buffer

	// Recorded state from the last call.
	executedModule string
	request        modules.Request
	help           model.HelpEnvelope
	syntaxErr      error
	syntaxSyntax   string
	doctorCalled   bool
	versionCalled  bool
	helpCalled     bool

	envelope render.MarkdownEnvelopeRenderer
}

func (s *stubRuntime) RunModule(_ context.Context, module string, req modules.Request) error {
	s.executedModule = module
	s.request = req
	return nil
}

func (s *stubRuntime) RenderHelp(_ context.Context, help model.HelpEnvelope) error {
	s.help = help
	s.helpCalled = true
	if s.stdout != nil {
		payload, err := s.envelope.RenderHelp(context.Background(), help)
		if err != nil {
			return err
		}
		_, _ = s.stdout.Write(payload)
	}
	return nil
}

func (s *stubRuntime) HandleSyntaxError(_ context.Context, err error, syntax string) error {
	s.syntaxErr = err
	s.syntaxSyntax = syntax
	if s.stdout != nil {
		message := "El comando es inválido."
		if err != nil && strings.TrimSpace(err.Error()) != "" {
			message = err.Error()
		}
		payload, renderErr := s.envelope.RenderFallback(context.Background(), model.FallbackEnvelope{
			Kind:       model.FallbackKindSyntax,
			Module:     "root",
			Title:      "dlexa",
			Message:    message,
			Syntax:     syntax,
			Suggestion: "Usá `--help` para ver sintaxis válida y ejemplos copiables.",
		})
		if renderErr != nil {
			return renderErr
		}
		_, _ = s.stdout.Write(payload)
	}
	return nil
}

func (s *stubRuntime) RunDoctor(_ context.Context) error {
	s.doctorCalled = true
	if s.stdout != nil {
		_, _ = fmt.Fprintf(s.stdout, "doctor: ok\n")
	}
	return nil
}

func (s *stubRuntime) PrintVersion() error {
	s.versionCalled = true
	if s.stdout != nil {
		_, _ = fmt.Fprintf(s.stdout, "dlexa dev\n")
	}
	return nil
}

// --- Routing tests ---

func TestRootCommand_QueryDefaultsToDPD(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"basto"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.executedModule != commandDPD {
		t.Fatalf(moduleWantDPDFormat, runtime.executedModule)
	}
	if runtime.request.Query != "basto" {
		t.Fatalf("query = %q, want basto", runtime.request.Query)
	}
}

func TestRootCommand_HelpFlag(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"--help"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if !runtime.helpCalled {
		t.Fatal("expected RenderHelp to be called")
	}
	if runtime.executedModule != "" {
		t.Fatal("RunModule should NOT be called for --help")
	}
}

func TestRootCommand_VersionFlag(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"--version"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if !runtime.versionCalled {
		t.Fatal("expected PrintVersion to be called")
	}
}

func TestRootCommand_DoctorFlag(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"--doctor"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if !runtime.doctorCalled {
		t.Fatal("expected RunDoctor to be called")
	}
}

func TestRootCommand_NoArgs(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if !runtime.helpCalled {
		t.Fatal("expected RenderHelp to be called for no args")
	}
}

func TestRootCommand_RootFormatFlagPropagates(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"basto", "--format", "json"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.executedModule != commandDPD {
		t.Fatalf(moduleWantDPDFormat, runtime.executedModule)
	}
	if runtime.request.Format != "json" {
		t.Fatalf("format = %q, want json", runtime.request.Format)
	}
}

func TestRootCommand_RootNoCacheFlagPropagates(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"basto", "--no-cache"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.executedModule != commandDPD {
		t.Fatalf(moduleWantDPDFormat, runtime.executedModule)
	}
	if !runtime.request.NoCache {
		t.Fatal("expected noCache=true")
	}
}

func TestRootCommand_RootRendersHelpEnvelopeWithExpectedContent(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"--help"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.help.Command != "dlexa" {
		t.Fatalf("help.Command = %q, want dlexa", runtime.help.Command)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("# Ayuda: dlexa")) {
		t.Fatalf("stdout = %q, missing help header", stdout.String())
	}
}
