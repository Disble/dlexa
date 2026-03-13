package query

import (
	"context"
	"errors"
	"testing"

	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/source"
)

func TestLookupPreservesDPDExtractFailures(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "dpd"}
	registry := &stubRegistry{sources: []source.Source{
		&stubSource{descriptor: descriptor, err: model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDExtractFailed,
			Message:  "extract failed",
			Source:   descriptor.Name,
			Severity: model.ProblemSeverityError,
		}, errors.New("shape changed"))},
	}}

	service := NewService(registry, &stubStore{})
	result, err := service.Lookup(context.Background(), model.LookupRequest{Query: "bien", Sources: []string{"dpd"}, NoCache: true})
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if len(result.Problems) != 1 {
		t.Fatalf("problems = %d, want 1", len(result.Problems))
	}
	if result.Problems[0].Code != model.ProblemCodeDPDExtractFailed {
		t.Fatalf("problem code = %q", result.Problems[0].Code)
	}
}
