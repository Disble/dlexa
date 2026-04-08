package main

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/modules"
)

func TestSearchCommandRoutesSemanticSearchModule(t *testing.T) {
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"search", "solo", "o", "sólo"}); err != nil {
		t.Fatalf("executeRootCommand() error = %v", err)
	}
	if runtime.executedModule != "search" {
		t.Fatalf("module = %q, want search", runtime.executedModule)
	}
	if !reflect.DeepEqual(runtime.request, modules.Request{Query: "solo o sólo", Args: []string{"solo", "o", "sólo"}}) {
		t.Fatalf("request = %#v", runtime.request)
	}
}

func TestSearchCommandRendersMarkdownHelp(t *testing.T) {
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"search", "--help"}); err != nil {
		t.Fatalf("executeRootCommand() error = %v", err)
	}
	if runtime.help.Command != "dlexa search" || !bytes.Contains(stdout.Bytes(), []byte("# Ayuda: dlexa search")) {
		t.Fatalf("help envelope = %#v stdout=%q", runtime.help, stdout.String())
	}
}

func TestSearchCommandTurnsMissingArgsIntoSyntaxFallback(t *testing.T) {
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"search"}); err != nil {
		t.Fatalf("executeRootCommand() error = %v", err)
	}
	if runtime.syntaxErr == nil || !bytes.Contains(stdout.Bytes(), []byte("Nivel 1 · Syntax")) {
		t.Fatalf("syntax err = %v stdout=%q", runtime.syntaxErr, stdout.String())
	}
	if got := runtime.syntaxSyntax; got != "dlexa search <consulta>" {
		t.Fatalf("syntax = %q, want %q", got, "dlexa search <consulta>")
	}
}
