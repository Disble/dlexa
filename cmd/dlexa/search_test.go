package main

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/modules"
)

type validateSourcesCase struct {
	name       string
	input      []string
	wantErr    bool
	wantErrSub string
}

type searchCommandSourceCase struct {
	name             string
	args             []string
	wantSources      []string
	wantSyntaxErr    bool
	wantSyntaxErrSub string
}

var validateSourcesCases = []validateSourcesCase{
	{name: "nil input", input: nil},
	{name: "dpd source", input: []string{"dpd"}},
	{name: "search source", input: []string{"search"}},
	{name: "dpd and search sources", input: []string{"dpd", "search"}},
	{name: "unknown source", input: []string{"unknown"}, wantErr: true, wantErrSub: "unknown"},
	{name: "mixed known and bad source", input: []string{"dpd", "bad"}, wantErr: true, wantErrSub: "bad"},
}

var searchCommandSourceCases = []searchCommandSourceCase{
	{name: "source flag routes dpd only", args: []string{"search", "--source", "dpd", "solo", "o", "sólo"}, wantSources: []string{"dpd"}},
	{name: "no source keeps federation", args: []string{"search", "solo", "o", "sólo"}},
	{name: "unknown source returns syntax fallback", args: []string{"search", "--source", "unknown", "solo"}, wantSyntaxErr: true, wantSyntaxErrSub: "unknown"},
}

func executeSearchArgs(t *testing.T, args []string) (*stubRuntime, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, args); err != nil {
		t.Fatalf(executeRootCommandErrorFormat, err)
	}
	return runtime, stdout, stderr
}

func TestSearchCommandRoutesSemanticSearchModule(t *testing.T) {
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandSearch, "solo", "o", "sólo"}); err != nil {
		t.Fatalf(executeRootCommandErrorFormat, err)
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
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandSearch, "--help"}); err != nil {
		t.Fatalf(executeRootCommandErrorFormat, err)
	}
	if runtime.help.Command != "dlexa search" || !bytes.Contains(stdout.Bytes(), []byte("# Ayuda: dlexa search")) {
		t.Fatalf("help envelope = %#v stdout=%q", runtime.help, stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("Descubrir entradas DPD, slugs editoriales y FAQs compatibles")) {
		t.Fatalf("stdout=%q, want capability-focused search guidance", stdout.String())
	}
	for _, want := range [][]byte{[]byte("## Qué podés hacer"), []byte("## Qué recibe"), []byte("## Guía para agentes y automatizaciones"), []byte("La salida distingue entre sugerencias ejecutables y guía diferida"), []byte("En JSON, inspeccioná cada candidato antes de automatizar")} {
		if !bytes.Contains(stdout.Bytes(), want) {
			t.Fatalf("stdout=%q, missing %q", stdout.String(), string(want))
		}
	}
	if bytes.Contains(stdout.Bytes(), []byte("Copiá el `next_command` sugerido")) {
		t.Fatalf("stdout=%q, should not instruct blindly copying next_command", stdout.String())
	}
}

func TestSearchCommandTurnsMissingArgsIntoSyntaxFallback(t *testing.T) {
	runtime := &stubRuntime{}
	stdout := &bytes.Buffer{}
	runtime.stdout = stdout
	stderr := &bytes.Buffer{}
	if err := executeRootCommand(context.Background(), runtime, stdout, stderr, []string{commandSearch}); err != nil {
		t.Fatalf(executeRootCommandErrorFormat, err)
	}
	if runtime.syntaxErr == nil || !bytes.Contains(stdout.Bytes(), []byte("Nivel 1 · Syntax")) {
		t.Fatalf("syntax err = %v stdout=%q", runtime.syntaxErr, stdout.String())
	}
	if got := runtime.syntaxSyntax; got != "dlexa search <consulta>" {
		t.Fatalf("syntax = %q, want %q", got, "dlexa search <consulta>")
	}
}

func TestValidateSources(t *testing.T) {
	for _, tc := range validateSourcesCases {
		t.Run(tc.name, func(t *testing.T) { assertValidateSourcesCase(t, tc) })
	}
}

func TestSearchCommandSourceFlag(t *testing.T) {
	for _, tc := range searchCommandSourceCases {
		t.Run(tc.name, func(t *testing.T) { assertSearchCommandSourceCase(t, tc) })
	}
}

func assertValidateSourcesCase(t *testing.T, tc validateSourcesCase) {
	t.Helper()

	err := validateSources(tc.input)
	if tc.wantErr {
		assertErrorContains(t, err, tc.wantErrSub, "validateSources(%v)", tc.input)
		return
	}
	if err != nil {
		t.Fatalf("validateSources(%v) error = %v, want nil", tc.input, err)
	}
}

func assertSearchCommandSourceCase(t *testing.T, tc searchCommandSourceCase) {
	t.Helper()

	runtime, _, _ := executeSearchArgs(t, tc.args)
	if tc.wantSyntaxErr {
		assertErrorContains(t, runtime.syntaxErr, tc.wantSyntaxErrSub, "syntax err")
		if runtime.executedModule != "" {
			t.Fatalf("executedModule = %q, want empty on invalid source", runtime.executedModule)
		}
		return
	}
	if runtime.executedModule != "search" {
		t.Fatalf("module = %q, want search", runtime.executedModule)
	}
	if !reflect.DeepEqual(runtime.request.Sources, tc.wantSources) {
		t.Fatalf("sources = %#v, want %#v", runtime.request.Sources, tc.wantSources)
	}
}

func assertErrorContains(t *testing.T, err error, wantSub string, label string, args ...any) {
	t.Helper()

	message := fmt.Sprintf(label, args...)
	if err == nil {
		t.Fatalf("%s error = nil, want error containing %q", message, wantSub)
	}
	if !strings.Contains(err.Error(), wantSub) {
		t.Fatalf("%s error = %q, want substring %q", message, err.Error(), wantSub)
	}
}
