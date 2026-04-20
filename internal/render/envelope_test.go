package render

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

var fallbackCases = []struct {
	name     string
	fallback model.FallbackEnvelope
	mustHave []string
}{
	{
		name:     "syntax fallback shows corrected syntax and help suggestion",
		fallback: model.FallbackEnvelope{Kind: model.FallbackKindSyntax, Module: "dpd", Title: "dlexa dpd", Message: "El comando es inválido.", Syntax: "dlexa dpd <termino>", Suggestion: "Usá `--help` para ver ejemplos."},
		mustHave: []string{"Nivel 1 · Syntax", "dlexa dpd <termino>", "--help"},
	},
	{
		name:     "not found fallback suggests search",
		fallback: model.FallbackEnvelope{Kind: model.FallbackKindNotFound, Module: "dpd", Query: "solo", Message: "No se encontró contenido en este módulo.", NextCommand: "dlexa search solo"},
		mustHave: []string{"Nivel 2 · Not Found", "dlexa search solo"},
	},
	{
		name:     "upstream fallback stops blind retries",
		fallback: model.FallbackEnvelope{Kind: model.FallbackKindUpstreamUnavailable, Module: "search", Query: "tilde", Suggestion: "NO reintentes en loop automático; abortá y reintentá más tarde."},
		mustHave: []string{"Nivel 3 · Upstream Unavailable", "NO reintentes"},
	},
	{
		name:     "rate limited fallback stops retries but names rate limit",
		fallback: model.FallbackEnvelope{Kind: model.FallbackKindRateLimited, Module: "search", Query: "tilde", Suggestion: "NO reintentes en loop automático; respetá el enfriamiento y reintentá más tarde."},
		mustHave: []string{"Nivel 3 · Rate Limited", "respetá el enfriamiento"},
	},
	{
		name:     "parse fallback requests maintenance",
		fallback: model.FallbackEnvelope{Kind: model.FallbackKindParseFailure, Module: "search", Query: "tilde", Suggestion: "Hace falta intervención humana de mantenimiento."},
		mustHave: []string{"Nivel 4 · Parse Failure", "intervención humana"},
	},
}

func TestEnvelopeRendererWrapsMarkdownAndBypassesJSON(t *testing.T) {
	renderer := NewEnvelopeRenderer()
	markdownPayload, err := renderer.RenderSuccess(context.Background(), model.Envelope{Module: "dpd", Title: "solo", Source: "Diccionario panhispánico de dudas", CacheState: "MISS", Format: "markdown"}, []byte("## Cuerpo\ncontenido"))
	if err != nil {
		t.Fatalf("RenderSuccess() markdown error = %v", err)
	}
	markdownText := string(markdownPayload)
	for _, want := range []string{"# [dlexa:dpd] solo", "*Fuente: Diccionario panhispánico de dudas | Caché: MISS*", "## Cuerpo"} {
		if !strings.Contains(markdownText, want) {
			t.Fatalf("markdown payload missing %q\n%s", want, markdownText)
		}
	}

	jsonPayload, err := renderer.RenderSuccess(context.Background(), model.Envelope{Module: "dpd", Title: "solo", Source: "Diccionario panhispánico de dudas", CacheState: "MISS", Format: "json"}, []byte(`{"unchanged":true}`))
	if err != nil {
		t.Fatalf("RenderSuccess() json error = %v", err)
	}
	if string(jsonPayload) != `{"unchanged":true}` {
		t.Fatalf("json payload = %s, want untouched body", string(jsonPayload))
	}
}

func TestEnvelopeRendererRendersMarkdownHelp(t *testing.T) {
	renderer := NewEnvelopeRenderer()
	payload, err := renderer.RenderHelp(context.Background(), model.HelpEnvelope{
		Command:      "dlexa search",
		Summary:      "Busca resultados lingüísticos y devuelve rutas de consulta.",
		Syntax:       "dlexa search <consulta>",
		Capabilities: []string{"Descubrir el módulo adecuado a partir de una consulta libre."},
		InputHints:   []string{"Recibe una consulta libre y puede devolver rutas ejecutables o diferidas."},
		Examples:     []string{"dlexa search solo o sólo", "dlexa search tilde en qué"},
		AgentNotes:   []string{"Leé la salida como router de comandos."},
		NextSteps:    []string{"Usá el comando sugerido para profundizar en el módulo correcto."},
	})
	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}
	text := string(payload)
	for _, want := range []string{"# Ayuda: dlexa search", "## Sintaxis", "## Qué podés hacer", "## Qué recibe", "## Guía para agentes y automatizaciones", "`dlexa search <consulta>`", "`dlexa search solo o sólo`", "Leé la salida como router de comandos.", "Usá el comando sugerido para profundizar en el módulo correcto."} {
		if !strings.Contains(text, want) {
			t.Fatalf("help payload missing %q\n%s", want, text)
		}
	}
}

func TestEnvelopeRendererRendersHelpSkippingBlankEntries(t *testing.T) {
	renderer := NewEnvelopeRenderer()
	payload, err := renderer.RenderHelp(context.Background(), model.HelpEnvelope{
		Examples:  []string{"", "  ", "dlexa search tilde"},
		NextSteps: []string{"", "Usá search."},
	})
	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}
	text := string(payload)
	for _, want := range []string{"# Ayuda: dlexa", "`dlexa search tilde`", "- Usá search."} {
		if !strings.Contains(text, want) {
			t.Fatalf("help payload missing %q\n%s", want, text)
		}
	}
	for _, forbidden := range []string{"- ``", "## Qué podés hacer", "## Qué recibe", "## Guía para agentes y automatizaciones"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("help payload contains forbidden %q\n%s", forbidden, text)
		}
	}
}

func TestEnvelopeRendererRendersFourLevelFallbacks(t *testing.T) {
	renderer := NewEnvelopeRenderer()
	for _, tc := range fallbackCases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := renderer.RenderFallback(context.Background(), tc.fallback)
			if err != nil {
				t.Fatalf("RenderFallback() error = %v", err)
			}
			text := string(payload)
			for _, want := range tc.mustHave {
				if !strings.Contains(text, want) {
					t.Fatalf("fallback payload missing %q\n%s", want, text)
				}
			}
		})
	}
}
