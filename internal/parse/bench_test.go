package parse

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

func loadBenchFixtureHTML(b *testing.B, name string) []byte {
	b.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", name+".html")) //nolint:gosec // G304: test fixture, path not from user input
	if err != nil {
		b.Fatalf("ReadFile() error = %v", err)
	}
	return body
}

func BenchmarkDPDParse(b *testing.B) {
	cases := []struct {
		name    string
		fixture string
	}{
		{"bien_single_article", "bien"},
		{"tilde_multi_article", "tilde"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			body := loadBenchFixtureHTML(b, tc.fixture)
			parser := NewDPDArticleParser()
			document := fetch.Document{
				URL:         "https://www.rae.es/dpd/" + tc.fixture,
				ContentType: "text/html; charset=utf-8",
				StatusCode:  200,
				Body:        body,
			}
			descriptor := model.SourceDescriptor{Name: "dpd", DisplayName: tc.fixture}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, err := parser.Parse(context.Background(), descriptor, document)
				if err != nil {
					b.Fatalf("Parse() error = %v", err)
				}
			}
		})
	}
}

func BenchmarkIsChallengePage(b *testing.B) {
	challengePrefix := "<html><title>Just a moment...</title><div>Cloudflare challenge</div>"
	nonChallengePrefix := "<html><head><title>DPD - bien</title></head><body><entry class=\"lex\">content</entry></body></html>"

	padString := func(prefix string, totalSize int) string {
		if len(prefix) >= totalSize {
			return prefix[:totalSize]
		}
		return prefix + strings.Repeat("x", totalSize-len(prefix))
	}

	cases := []struct {
		name string
		body string
	}{
		{"challenge_1KB", padString(challengePrefix, 1024)},
		{"challenge_500KB", padString(challengePrefix, 500*1024)},
		{"non_challenge_1KB", padString(nonChallengePrefix, 1024)},
		{"non_challenge_500KB", padString(nonChallengePrefix, 500*1024)},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				isChallengePage(tc.body)
			}
		})
	}
}

func BenchmarkPreserveSemanticSpans(b *testing.B) {
	// Build fixtures with varying tag counts but similar text size (~5KB).

	buildFixture := func(allowedTags, unknownTags int) string {
		var builder strings.Builder
		for i := 0; i < allowedTags; i++ {
			builder.WriteString(`<span class="ejemplo">ejemplo texto de prueba</span> `)
			builder.WriteString(`<em>enfasis texto</em> `)
			builder.WriteString(`<dfn>definicion</dfn> `)
		}
		for i := 0; i < unknownTags; i++ {
			builder.WriteString(`<span class="unknown-class">texto desconocido</span> `)
			builder.WriteString(`<div>bloque eliminable</div> `)
		}
		// Pad to ~5KB if needed.
		for builder.Len() < 5000 {
			builder.WriteString("Texto de relleno para alcanzar tamaño objetivo. ")
		}
		return builder.String()
	}

	cases := []struct {
		name  string
		input string
	}{
		{"small_5_tags", buildFixture(3, 2)},
		{"medium_20_tags", buildFixture(5, 15)},
		{"large_50_tags", buildFixture(10, 40)},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				preserveSemanticSpans(tc.input)
			}
		})
	}
}
