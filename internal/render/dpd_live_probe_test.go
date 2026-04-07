package render

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
)

func TestDPDLiveProbeBienDriftInvariants(t *testing.T) {
	if os.Getenv("DLEXA_LIVE_DPD_PROBE") == "" {
		t.Skip("set DLEXA_LIVE_DPD_PROBE=1 to run live DPD drift probe")
	}

	runtimeConfig := config.DefaultRuntimeConfig()
	document, err := fetch.NewDPDFetcher(runtimeConfig.DPD.BaseURL, runtimeConfig.DPD.Timeout, runtimeConfig.DPD.UserAgent).Fetch(
		context.Background(),
		fetch.Request{Query: "bien", Source: model.SourceDescriptor{Name: "dpd", DisplayName: "Diccionario panhispánico de dudas"}},
	)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	parsed, _, err := parse.NewDPDArticleParser().Parse(
		context.Background(),
		model.SourceDescriptor{Name: "dpd", DisplayName: "Diccionario panhispánico de dudas"},
		document,
	)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	normalized, err := normalize.NewDPDNormalizer().Normalize(
		context.Background(),
		model.SourceDescriptor{Name: "dpd", DisplayName: "Diccionario panhispánico de dudas"},
		parsed,
	)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(normalized.Entries) == 0 {
		t.Fatal("normalized entries = 0, want live bien article")
	}

	payload, err := NewMarkdownRenderer().Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "bien", Format: "markdown"},
		Entries: normalized.Entries,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	text := string(payload)
	for _, want := range []string{
		"# bien",
		"Diccionario panhispánico de dudas",
		"2.ª edición",
		"5. bien que.",
		"6. más bien.",
		"7. si bien.",
		"Source: Real Academia Española y Asociación de Academias de la Lengua Española",
		"URL: https://www.rae.es/dpd/bien",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("live probe output missing %q\n%s", want, text)
		}
	}
	for _, forbidden := range []string{"Cloudflare", "Just a moment", "newsletter", "compartir"} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {
			t.Fatalf("live probe output leaked upstream shell/challenge marker %q\n%s", forbidden, text)
		}
	}
}
