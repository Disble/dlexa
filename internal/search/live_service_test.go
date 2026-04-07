package search

import (
	"context"
	"errors"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
	"github.com/Disble/dlexa/internal/testutil"
)

func TestServicePreservesLiveParseFailures(t *testing.T) {
	wantErr := model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchParseFailed, Message: "live search markup changed", Source: "search", Severity: model.ProblemSeverityError}, errors.New("markup changed"))
	service := NewService(model.SourceDescriptor{Name: "search"}, &stubFetcher{document: fetch.Document{Body: []byte("<html></html>")}}, &stubParser{err: wantErr}, &stubNormalizer{}, &stubSearchStore{})

	_, err := service.Search(context.Background(), model.SearchRequest{Query: testutil.LiveSearchQuery})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Search() error = %v, want %v", err, wantErr)
	}
}

func TestServicePreservesLiveNormalizeFailures(t *testing.T) {
	wantErr := model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchNormalizeFailed, Message: "normalize failed", Source: "search", Severity: model.ProblemSeverityError}, errors.New("normalize failed"))
	service := NewService(model.SourceDescriptor{Name: "search"}, &stubFetcher{document: fetch.Document{Body: []byte("<html></html>")}}, &stubParser{records: []parse.ParsedSearchRecord{{Title: "bien", URL: testutil.LiveSearchDPDURL}}}, &stubNormalizer{err: wantErr}, &stubSearchStore{})

	_, err := service.Search(context.Background(), model.SearchRequest{Query: testutil.LiveSearchQuery})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Search() error = %v, want %v", err, wantErr)
	}
}
