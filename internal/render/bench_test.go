// Package render benchmarks for render output formats.
package render

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
)

func loadBenchFixtureHTML(b *testing.B, name string) []byte {
	b.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", name+".html")) //nolint:gosec // G304: test fixture, path not from user input
	if err != nil {
		b.Fatalf("ReadFile() error = %v", err)
	}
	return body
}

func benchParseNormalize(b *testing.B, term string) []model.Entry {
	b.Helper()
	body := loadBenchFixtureHTML(b, term)
	parser := parse.NewDPDArticleParser()
	parsed, _, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd", DisplayName: term}, fetch.Document{
		URL:         "https://www.rae.es/dpd/" + term,
		ContentType: "text/html; charset=utf-8",
		StatusCode:  200,
		Body:        body,
	})
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}

	normalizer := normalize.NewDPDNormalizer()
	normalized, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, parsed)
	if err != nil {
		b.Fatalf("Normalize() error = %v", err)
	}
	return normalized.Entries
}

func BenchmarkMarkdownRender(b *testing.B) {
	cases := []struct {
		name    string
		fixture string
	}{
		{"bien_single_article", "bien"},
		{"tilde_multi_article", "tilde"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			entries := benchParseNormalize(b, tc.fixture)
			renderer := NewMarkdownRenderer()
			result := model.LookupResult{
				Request: model.LookupRequest{Query: tc.fixture, Format: "markdown"},
				Entries: entries,
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := renderer.Render(context.Background(), result)
				if err != nil {
					b.Fatalf("Render() error = %v", err)
				}
			}
		})
	}
}
