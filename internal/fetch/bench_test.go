// Package fetch benchmarks for HTTP fetcher and challenge detection.
package fetch

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

// fixtureSmallHTML is a minimal DPD-like article page.
const fixtureSmallHTML = `<!DOCTYPE html>
<html lang="es">
<head><title>DPD - bien</title></head>
<body>
<div id="content">
<entry class="lex" id="bien" header="bien">
<header class="lex">bien</header>
<p n="1n"><span class="enum">1.</span> Como adverbio de modo significa <dfn>'correcta y adecuadamente'</dfn>: <span class="ejemplo">Cierra bien la ventana, por favor</span>.</p>
</entry>
</div>
<p class="o">Real Academia Española y Asociación de Academias de la Lengua Española: <i>Diccionario panhispánico de dudas (DPD)</i> [en línea], https://www.rae.es/dpd/bien, 2.ª edición. [Consulta: 10/03/2026].</p>
<span class='aviso-edicion'><span>2.ª edición</span></span>
</body>
</html>`

// fixtureLargeHTML simulates a large DPD page (~50KB) by repeating sections.
var fixtureLargeHTML = func() string {
	var builder strings.Builder
	builder.WriteString(`<!DOCTYPE html><html lang="es"><head><title>DPD - ser</title></head><body><div id="content"><entry class="lex" id="ser" header="ser"><header class="lex">ser</header>`)
	for i := 0; i < 100; i++ {
		builder.WriteString(`<p n="1n"><span class="enum">1.</span> Verbo irregular. <span class="ejemplo">El sol es una estrella</span>. Su comparativo es <em><span class="ment">mejor</span></em>. Considérese el uso de <dfn>'existir'</dfn> en contextos formales (→ <a href="ser#S123">2</a>).</p>`)
	}
	builder.WriteString(`</entry></div>`)
	builder.WriteString(`<p class="o">Real Academia Española: <i>Diccionario panhispánico de dudas</i>, https://www.rae.es/dpd/ser. [Consulta: 10/03/2026].</p>`)
	builder.WriteString(`<span class='aviso-edicion'><span>2.ª edición</span></span></body></html>`)
	return builder.String()
}()

func BenchmarkDPDFetch(b *testing.B) {
	cases := []struct {
		name     string
		bodyHTML string
	}{
		{"small_article", fixtureSmallHTML},
		{"large_article", fixtureLargeHTML},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, tc.bodyHTML)
			}))
			defer server.Close()

			fetcher := NewDPDFetcher(server.URL+"/dpd", 5*time.Second, "dlexa-bench")
			request := Request{
				Query:  "bien",
				Source: model.SourceDescriptor{Name: "dpd"},
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := fetcher.Fetch(context.Background(), request)
				if err != nil {
					b.Fatalf("Fetch() error = %v", err)
				}
			}
		})
	}
}

func BenchmarkIsChallengeBody(b *testing.B) {
	challengePrefix := []byte("<html><title>Just a moment...</title><div>Cloudflare challenge</div>")
	nonChallengePrefix := []byte("<html><head><title>DPD - bien</title></head><body><entry class=\"lex\">content</entry></body></html>")

	padBody := func(prefix []byte, totalSize int) []byte {
		body := make([]byte, totalSize)
		copy(body, prefix)
		for i := len(prefix); i < totalSize; i++ {
			body[i] = 'x'
		}
		return body
	}

	cases := []struct {
		name string
		body []byte
	}{
		{"challenge_1KB", padBody(challengePrefix, 1024)},
		{"challenge_100KB", padBody(challengePrefix, 100*1024)},
		{"challenge_500KB", padBody(challengePrefix, 500*1024)},
		{"non_challenge_1KB", padBody(nonChallengePrefix, 1024)},
		{"non_challenge_100KB", padBody(nonChallengePrefix, 100*1024)},
		{"non_challenge_500KB", padBody(nonChallengePrefix, 500*1024)},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				isChallengeBody(tc.body)
			}
		})
	}
}
