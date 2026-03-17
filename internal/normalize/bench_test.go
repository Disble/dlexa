package normalize

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
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

func parseBenchFixture(b *testing.B, name string) parse.Result {
	b.Helper()
	body := loadBenchFixtureHTML(b, name)
	parser := parse.NewDPDArticleParser()
	result, _, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd", DisplayName: name}, fetch.Document{
		URL:         "https://www.rae.es/dpd/" + name,
		ContentType: "text/html; charset=utf-8",
		StatusCode:  200,
		Body:        body,
	})
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}
	return result
}

func BenchmarkDPDNormalize(b *testing.B) {
	cases := []struct {
		name    string
		fixture string
	}{
		{"bien_single_article", "bien"},
		{"tilde_multi_article", "tilde"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			parsed := parseBenchFixture(b, tc.fixture)
			normalizer := NewDPDNormalizer()
			descriptor := model.SourceDescriptor{Name: "dpd"}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := normalizer.Normalize(context.Background(), descriptor, parsed)
				if err != nil {
					b.Fatalf("Normalize() error = %v", err)
				}
			}
		})
	}
}
