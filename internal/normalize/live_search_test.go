package normalize

import (
	"context"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
	"github.com/Disble/dlexa/internal/testutil"
)

func TestLiveSearchNormalizerBuildsCuratedCandidates(t *testing.T) {
	normalizer := NewLiveSearchNormalizer()
	candidates, warnings, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: "search"}, []parse.ParsedSearchRecord{{Title: testutil.LiveSearchDPDTitle, Snippet: testutil.LiveSearchDPDSnippet, URL: testutil.LiveSearchDPDURL}, {Title: testutil.LiveSearchUnknownTitle, Snippet: testutil.LiveSearchUnknownSnippet, URL: testutil.LiveSearchUnknownURL}})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if len(candidates) != 2 {
		t.Fatalf("candidates len = %d, want 2", len(candidates))
	}
	if got := candidates[0]; got.Title != testutil.LiveSearchDPDTitle || got.Snippet != testutil.LiveSearchDPDSnippet || got.URL != testutil.LiveSearchDPDURL || got.DisplayText != testutil.LiveSearchDPDTitle {
		t.Fatalf("first candidate = %#v", got)
	}
	if got := candidates[1].SourceHint; got != "RAE" {
		t.Fatalf("SourceHint = %q, want RAE", got)
	}
}

func TestLiveSearchNormalizerRejectsEntirelyUnusableRecords(t *testing.T) {
	_, _, err := NewLiveSearchNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "search"}, []parse.ParsedSearchRecord{{Title: "", Snippet: "", URL: ""}})
	if err == nil {
		t.Fatal("Normalize() error = nil, want problem")
	}
	problem, ok := model.AsProblem(err)
	if !ok {
		t.Fatalf("Normalize() error = %T, want problem", err)
	}
	if problem.Code != model.ProblemCodeDPDSearchNormalizeFailed {
		t.Fatalf("problem code = %q, want %q", problem.Code, model.ProblemCodeDPDSearchNormalizeFailed)
	}
}
