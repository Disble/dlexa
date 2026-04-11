package modules

import (
	"context"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestRequestResponseContractsExposeSharedModuleData(t *testing.T) {
	t.Helper()
	req := Request{Query: "solo", Format: "markdown", NoCache: true, Args: []string{"solo"}, Sources: []string{"dpd"}}
	fallback := &model.FallbackEnvelope{Kind: model.FallbackKindNotFound, Module: "dpd", Query: "solo", NextCommand: "dlexa search solo"}
	resp := Response{Title: "solo", Source: "Diccionario panhispánico de dudas", CacheState: "MISS", Format: "markdown", Body: []byte("body"), Fallback: fallback}

	if req.Query != "solo" || req.Format != "markdown" || !req.NoCache {
		t.Fatalf("request = %#v", req)
	}
	if resp.Fallback == nil || resp.Fallback.Kind != model.FallbackKindNotFound {
		t.Fatalf("response fallback = %#v", resp.Fallback)
	}
	if CacheState(true) != "HIT" || CacheState(false) != "MISS" {
		t.Fatalf("CacheState() returned unstable markers")
	}

	registry := NewRegistry(stubModule{name: "Diccionario panhispánico de dudas", command: "dpd", response: resp})
	module, ok := registry.Module("dpd")
	if !ok {
		t.Fatal("Module() did not resolve registered module")
	}
	got, err := module.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.Title != resp.Title || got.Source != resp.Source {
		t.Fatalf("Execute() response = %#v, want %#v", got, resp)
	}
}

func TestFallbackFromErrorDistinguishesRateLimitedFromGenericUpstream(t *testing.T) {
	rateLimited := FallbackFromError("search", "solo", "markdown", model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchRateLimited, Message: "limited", Severity: model.ProblemSeverityError}, nil))
	if rateLimited == nil || rateLimited.Kind != model.FallbackKindRateLimited {
		t.Fatalf("rateLimited fallback = %#v, want rate_limited", rateLimited)
	}
	if rateLimited.Message != "La fuente externa pidió frenar temporalmente por rate limit." {
		t.Fatalf("rateLimited message = %q", rateLimited.Message)
	}

	upstream := FallbackFromError("search", "solo", "markdown", model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "down", Severity: model.ProblemSeverityError}, nil))
	if upstream == nil || upstream.Kind != model.FallbackKindUpstreamUnavailable {
		t.Fatalf("upstream fallback = %#v, want upstream_unavailable", upstream)
	}
}

type stubModule struct {
	name     string
	command  string
	response Response
}

func (s stubModule) Name() string                                       { return s.name }
func (s stubModule) Command() string                                    { return s.command }
func (s stubModule) Execute(context.Context, Request) (Response, error) { return s.response, nil }
