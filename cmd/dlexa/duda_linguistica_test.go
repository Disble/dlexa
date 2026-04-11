package main

import (
	"bytes"
	"context"
	"testing"
)

func TestDudaLinguisticaCommandRoutesModule(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandDudaLinguistica, "cuando-se-escriben-con-tilde-los-adverbios-en-mente"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.executedModule != commandDudaLinguistica {
		t.Fatalf("module = %q, want %q", runtime.executedModule, commandDudaLinguistica)
	}
	if runtime.request.Query != "cuando-se-escriben-con-tilde-los-adverbios-en-mente" {
		t.Fatalf("query = %q, want slug", runtime.request.Query)
	}
}

func TestDudaLinguisticaCommandRendersHelp(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandDudaLinguistica, "--help"}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.help.Command != helpCommandDudaLinguistica {
		t.Fatalf("help.Command = %q, want %q", runtime.help.Command, helpCommandDudaLinguistica)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("# Ayuda: "+helpCommandDudaLinguistica)) {
		t.Fatalf("stdout = %q, missing help header", stdout.String())
	}
}

func TestDudaLinguisticaCommandTurnsMissingArgsIntoSyntaxFallback(t *testing.T) {
	runtime := &stubRuntime{}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.stdout = stdout

	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandDudaLinguistica}); err != nil {
		t.Fatalf(unexpectedErrorFormat, err)
	}
	if runtime.syntaxErr == nil {
		t.Fatal("expected syntax error")
	}
	if got := runtime.syntaxSyntax; got != syntaxDudaLinguistica {
		t.Fatalf("syntax = %q, want %q", got, syntaxDudaLinguistica)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("Nivel 1 · Syntax")) {
		t.Fatalf("stdout = %q, want syntax fallback", stdout.String())
	}
}
