package main

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/modules"
)

var validateSourcesCases = []struct {
	name       string
	input      []string
	wantErr    bool
	wantErrSub string
}{
	{name: "nil input", input: nil},
	{name: "dpd source", input: []string{"dpd"}},
	{name: "search source", input: []string{"search"}},
	{name: "dpd and search sources", input: []string{"dpd", "search"}},
	{name: "unknown source", input: []string{"unknown"}, wantErr: true, wantErrSub: "unknown"},
	{name: "mixed known and bad source", input: []string{"dpd", "bad"}, wantErr: true, wantErrSub: "bad"},
}

var searchCommandSourceCases = []struct {
	name             string
	args             []string
	wantSources      []string
	wantSyntaxErr    bool
	wantSyntaxErrSub string
}{
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
		t.Fatalf("executeRootCommand() error = %v", err)
	}
	return runtime, stdout, stderr
}

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
	if !bytes.Contains(stdout.Bytes(), []byte("algunas rutas ya son ejecutables")) {
		t.Fatalf("stdout=%q, want mixed executable/deferred guidance disclaimer", stdout.String())
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

func TestValidateSources(t *testing.T) {
	for _, tc := range validateSourcesCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSources(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("validateSources(%v) error = nil, want error containing %q", tc.input, tc.wantErrSub)
				}
				if !strings.Contains(err.Error(), tc.wantErrSub) {
					t.Fatalf("validateSources(%v) error = %q, want substring %q", tc.input, err.Error(), tc.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateSources(%v) error = %v, want nil", tc.input, err)
			}
		})
	}
}

func TestSearchCommandSourceFlag(t *testing.T) {
	for _, tc := range searchCommandSourceCases {
		t.Run(tc.name, func(t *testing.T) {
			runtime, _, _ := executeSearchArgs(t, tc.args)
			if tc.wantSyntaxErr {
				if runtime.syntaxErr == nil {
					t.Fatalf("syntax err = nil, want error containing %q", tc.wantSyntaxErrSub)
				}
				if !strings.Contains(runtime.syntaxErr.Error(), tc.wantSyntaxErrSub) {
					t.Fatalf("syntax err = %q, want substring %q", runtime.syntaxErr.Error(), tc.wantSyntaxErrSub)
				}
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
		})
	}
}
