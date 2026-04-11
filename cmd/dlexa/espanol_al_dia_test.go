package main

import (
	"bytes"
	"context"
	"testing"
)

func TestEspanolAlDiaCommandRoutesModule(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandEspanolAlDia, "el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.executedModule != commandEspanolAlDia {
		t.Fatalf("module = %q, want %q", runtime.executedModule, commandEspanolAlDia)
	}
	if runtime.request.Query != "el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde" {
		t.Fatalf("query = %q, want slug", runtime.request.Query)
	}
}

func TestEspanolAlDiaCommandRendersHelp(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandEspanolAlDia, "--help"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.help.Command != helpCommandEspanolAlDia {
		t.Fatalf("help.Command = %q, want %q", runtime.help.Command, helpCommandEspanolAlDia)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("# Ayuda: "+helpCommandEspanolAlDia)) {
		t.Fatalf("stdout = %q, missing help header", stdout.String())
	}
}

func TestEspanolAlDiaCommandTurnsMissingArgsIntoSyntaxFallback(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandEspanolAlDia}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.syntaxErr == nil {
		t.Fatal("expected syntax error")
	}
	if got := runtime.syntaxSyntax; got != syntaxEspanolAlDia {
		t.Fatalf("syntax = %q, want %q", got, syntaxEspanolAlDia)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("Nivel 1 · Syntax")) {
		t.Fatalf("stdout = %q, want syntax fallback", stdout.String())
	}
}
