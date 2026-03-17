package parse

import (
	"context"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

func TestDPDSearchParserDecodesJSONArrayAndSplitsOnFirstPipe(t *testing.T) {
	parser := NewDPDSearchParser()
	records, warnings, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd"}, fetch.Document{Body: []byte(`["<em>Abu Dhabi</em>|Abu Dabi","guion|guion|extra","malformed"]`)})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if len(records) != 2 {
		t.Fatalf("records = %#v", records)
	}
	if records[0].RawLabelHTML != "<em>Abu Dhabi</em>" || records[0].ArticleKey != "Abu Dabi" {
		t.Fatalf("first record = %#v", records[0])
	}
	if records[1].RawLabelHTML != "guion" || records[1].ArticleKey != "guion|extra" {
		t.Fatalf("second record = %#v", records[1])
	}
}

func TestDPDSearchParserRejectsNonJSONAndEntirelyUnusablePayloads(t *testing.T) {
	tests := []struct {
		name string
		body string
		code string
	}{
		{name: "non json body", body: `<html>not json</html>`, code: model.ProblemCodeDPDSearchParseFailed},
		{name: "all malformed items", body: `["malformed","still-bad"]`, code: model.ProblemCodeDPDSearchParseFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := NewDPDSearchParser().Parse(context.Background(), model.SourceDescriptor{Name: "dpd"}, fetch.Document{Body: []byte(tt.body)})
			if err == nil {
				t.Fatal("Parse() error = nil, want problem")
			}
			problem, ok := model.AsProblem(err)
			if !ok {
				t.Fatalf("Parse() error = %T, want problem", err)
			}
			if problem.Code != tt.code {
				t.Fatalf("problem code = %q, want %q", problem.Code, tt.code)
			}
		})
	}
}
