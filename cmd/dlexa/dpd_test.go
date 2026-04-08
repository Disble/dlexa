package main

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/modules"
)

func TestDPDCommandRoutesExplicitDPDModule(t *testing.T) {
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"dpd", "solo", "--format", "json", "--no-cache"}); err != nil {
		t.Fatalf("executeRootCommand() error = %v", err)
	}
	if runtime.executedModule != "dpd" {
		t.Fatalf("module = %q, want dpd", runtime.executedModule)
	}
	if !reflect.DeepEqual(runtime.request, modules.Request{Query: "solo", Format: "json", NoCache: true, Args: []string{"solo"}}) {
		t.Fatalf("request = %#v", runtime.request)
	}
}

func TestDPDCommandRendersMarkdownHelp(t *testing.T) {
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"dpd", "--help"}); err != nil {
		t.Fatalf("executeRootCommand() error = %v", err)
	}
	if runtime.help.Command != "dlexa dpd" || !bytes.Contains(stdout.Bytes(), []byte("# Ayuda: dlexa dpd")) {
		t.Fatalf("help envelope = %#v stdout=%q", runtime.help, stdout.String())
	}
}

func TestDPDCommandTurnsMissingArgsIntoSyntaxFallback(t *testing.T) {
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"dpd"}); err != nil {
		t.Fatalf("executeRootCommand() error = %v", err)
	}
	if runtime.syntaxErr == nil || !bytes.Contains(stdout.Bytes(), []byte("Nivel 1 · Syntax")) {
		t.Fatalf("syntax err = %v stdout=%q", runtime.syntaxErr, stdout.String())
	}
	if got := runtime.syntaxSyntax; got != "dlexa dpd <termino>" {
		t.Fatalf("syntax = %q, want %q", got, "dlexa dpd <termino>")
	}
}
