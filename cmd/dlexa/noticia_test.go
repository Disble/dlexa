package main

import (
	"bytes"
	"context"
	"testing"
)

func TestNoticiaCommandRoutesModule(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandNoticia, "preguntas-frecuentes-tilde-en-las-mayusculas"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.executedModule != "noticia" {
		t.Fatalf("module = %q, want noticia", runtime.executedModule)
	}
	if runtime.request.Query != "preguntas-frecuentes-tilde-en-las-mayusculas" {
		t.Fatalf("query = %q, want slug", runtime.request.Query)
	}
}

func TestNoticiaCommandRendersHelp(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandNoticia, "--help"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.help.Command != "dlexa noticia" {
		t.Fatalf("help.Command = %q, want dlexa noticia", runtime.help.Command)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("# Ayuda: dlexa noticia")) {
		t.Fatalf("stdout = %q, missing help header", stdout.String())
	}
}

func TestNoticiaCommandTurnsMissingArgsIntoSyntaxFallback(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandNoticia}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.syntaxErr == nil {
		t.Fatal("expected syntax error")
	}
	if got := runtime.syntaxSyntax; got != syntaxNoticia {
		t.Fatalf("syntax = %q, want %q", got, syntaxNoticia)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("Nivel 1 · Syntax")) {
		t.Fatalf("stdout = %q, want syntax fallback", stdout.String())
	}
}
