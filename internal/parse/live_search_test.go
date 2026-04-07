package parse

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/testutil"
)

func TestLiveSearchParserExtractsCuratedRecordsFromSearchMarkup(t *testing.T) {
	parser := NewLiveSearchParser()
	records, warnings, err := parser.Parse(
		context.Background(),
		model.SourceDescriptor{Name: "search"},
		fetch.Document{
			URL:  "https://www.rae.es/search/node?keys=solo+o+solo",
			Body: mustReadFixture(t, testutil.LiveSearchResultsFixture),
		},
	)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if len(records) != 4 {
		t.Fatalf("records len = %d, want 4", len(records))
	}
	if got := records[0]; got.Title != "El adverbio «solo» y los pronombres demostrativos, sin tilde" || got.URL != "https://www.rae.es/espanol-al-dia/el-adverbio-solo-y-los-pronombres-demostrativos-sin-tilde" {
		t.Fatalf("first record = %#v", got)
	}
	if got := records[1].Snippet; got == "" {
		t.Fatal("second record snippet = empty, want normalized snippet text")
	}
}

func TestLiveSearchParserAllowsEmptySearchPayload(t *testing.T) {
	records, warnings, err := NewLiveSearchParser().Parse(
		context.Background(),
		model.SourceDescriptor{Name: "search"},
		fetch.Document{URL: "https://www.rae.es/search/node?keys=zzz", Body: mustReadFixture(t, testutil.LiveSearchEmptyFixture)},
	)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if len(records) != 0 {
		t.Fatalf("records = %#v, want empty", records)
	}
}

func TestLiveSearchParserReportsBrokenMarkupExplicitly(t *testing.T) {
	_, _, err := NewLiveSearchParser().Parse(
		context.Background(),
		model.SourceDescriptor{Name: "search"},
		fetch.Document{URL: "https://www.rae.es/search/node?keys=solo+o+solo", Body: mustReadFixture(t, testutil.LiveSearchBrokenFixture)},
	)
	if err == nil {
		t.Fatal("Parse() error = nil, want problem")
	}
	problem, ok := model.AsProblem(err)
	if !ok {
		t.Fatalf("Parse() error = %T, want problem", err)
	}
	if problem.Code != model.ProblemCodeDPDSearchParseFailed {
		t.Fatalf("problem code = %q, want %q", problem.Code, model.ProblemCodeDPDSearchParseFailed)
	}
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", filepath.Base(name)))
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	return raw
}
