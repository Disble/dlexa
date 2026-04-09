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

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"espanol-al-dia", "el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runtime.executedModule != "espanol-al-dia" {
		t.Fatalf("module = %q, want espanol-al-dia", runtime.executedModule)
	}
	if runtime.request.Query != "el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde" {
		t.Fatalf("query = %q, want slug", runtime.request.Query)
	}
}

func TestEspanolAlDiaCommandRendersHelp(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"espanol-al-dia", "--help"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runtime.help.Command != "dlexa espanol-al-dia" {
		t.Fatalf("help.Command = %q, want dlexa espanol-al-dia", runtime.help.Command)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("# Ayuda: dlexa espanol-al-dia")) {
		t.Fatalf("stdout = %q, missing help header", stdout.String())
	}
}

func TestEspanolAlDiaCommandTurnsMissingArgsIntoSyntaxFallback(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{"espanol-al-dia"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runtime.syntaxErr == nil {
		t.Fatal("expected syntax error")
	}
	if got := runtime.syntaxSyntax; got != "dlexa espanol-al-dia <slug>" {
		t.Fatalf("syntax = %q, want dlexa espanol-al-dia <slug>", got)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("Nivel 1 · Syntax")) {
		t.Fatalf("stdout = %q, want syntax fallback", stdout.String())
	}
}
