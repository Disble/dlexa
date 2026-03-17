package search

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

func TestServiceReturnsCachedResultAndRefreshesRequestFields(t *testing.T) {
	request := model.SearchRequest{Query: " Abu Dhabi ", Format: "json"}
	stored := model.SearchResult{
		Request:    model.SearchRequest{Query: "Abu Dhabi", Format: "markdown"},
		Candidates: []model.SearchCandidate{{RawLabelHTML: "<em>Abu Dhabi</em>", DisplayText: "Abu Dhabi", ArticleKey: "Abu Dabi"}},
	}

	service := NewService(
		model.SourceDescriptor{Name: "dpd"},
		nil,
		nil,
		nil,
		&stubSearchStore{getResult: stored, getOK: true},
	)

	result, err := service.Search(context.Background(), request)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if !result.CacheHit {
		t.Fatal("CacheHit = false, want true")
	}

	if !reflect.DeepEqual(result.Request, request) {
		t.Fatalf("Request = %#v, want %#v", result.Request, request)
	}

	if len(result.Candidates) != 1 || result.Candidates[0].ArticleKey != "Abu Dabi" {
		t.Fatalf("Candidates = %#v", result.Candidates)
	}
}

func TestServiceFetchesParsesNormalizesAndCachesSearchResults(t *testing.T) {
	fixedNow := time.Date(2026, time.March, 17, 21, 0, 0, 0, time.UTC)
	request := model.SearchRequest{Query: "guion", Format: "markdown"}
	store := &stubSearchStore{}
	fetcher := &stubFetcher{document: fetch.Document{Body: []byte(`["guion<sup>1</sup>|guion","<span class=\"vers\">guion<sup>2</sup></span>|guion"]`)}}
	parser := &stubParser{records: []parse.ParsedSearchRecord{{RawLabelHTML: "guion<sup>1</sup>", ArticleKey: "guion"}, {RawLabelHTML: `<span class="vers">guion<sup>2</sup></span>`, ArticleKey: "guion"}}}
	normalizer := &stubNormalizer{candidates: []model.SearchCandidate{{RawLabelHTML: "guion<sup>1</sup>", DisplayText: "guion1", ArticleKey: "guion"}, {RawLabelHTML: `<span class="vers">guion<sup>2</sup></span>`, DisplayText: "var. guion2", ArticleKey: "guion"}}}

	service := NewService(model.SourceDescriptor{Name: "dpd"}, fetcher, parser, normalizer, store)
	service.now = func() time.Time { return fixedNow }

	result, err := service.Search(context.Background(), request)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if fetcher.request.Query != request.Query {
		t.Fatalf("Fetch query = %q, want %q", fetcher.request.Query, request.Query)
	}

	if !reflect.DeepEqual(parser.document, fetcher.document) {
		t.Fatalf("parser document = %#v, want %#v", parser.document, fetcher.document)
	}

	if len(result.Candidates) != 2 {
		t.Fatalf("Candidates len = %d, want 2", len(result.Candidates))
	}

	if result.GeneratedAt != fixedNow {
		t.Fatalf("GeneratedAt = %v, want %v", result.GeneratedAt, fixedNow)
	}

	if store.setCalls != 1 {
		t.Fatalf("Set() calls = %d, want 1", store.setCalls)
	}

	if store.setResult.Request.Format != "" {
		t.Fatalf("cached Request.Format = %q, want empty for format-neutral cache data", store.setResult.Request.Format)
	}
}

func TestServiceTreatsEmptyCandidateSetAsSuccessfulSearch(t *testing.T) {
	service := NewService(
		model.SourceDescriptor{Name: "dpd"},
		&stubFetcher{document: fetch.Document{Body: []byte(`[]`)}},
		&stubParser{},
		&stubNormalizer{},
		&stubSearchStore{},
	)

	result, err := service.Search(context.Background(), model.SearchRequest{Query: "no existe"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(result.Candidates) != 0 {
		t.Fatalf("Candidates = %#v, want empty", result.Candidates)
	}
}

func TestServicePreservesUpstreamFailures(t *testing.T) {
	wantErr := model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "search unavailable", Source: "dpd", Severity: model.ProblemSeverityError}, errors.New("timeout"))
	service := NewService(model.SourceDescriptor{Name: "dpd"}, &stubFetcher{err: wantErr}, &stubParser{}, &stubNormalizer{}, &stubSearchStore{})

	_, err := service.Search(context.Background(), model.SearchRequest{Query: "tilde"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Search() error = %v, want %v", err, wantErr)
	}
}

type stubSearchStore struct {
	getResult model.SearchResult
	getOK     bool
	getErr    error
	setErr    error
	setCalls  int
	setResult model.SearchResult
	setKey    string
}

func (s *stubSearchStore) Get(_ context.Context, _ string) (model.SearchResult, bool, error) {
	return s.getResult, s.getOK, s.getErr
}

func (s *stubSearchStore) Set(_ context.Context, key string, result model.SearchResult) error {
	s.setCalls++
	s.setKey = key
	s.setResult = result
	return s.setErr
}

type stubFetcher struct {
	request  fetch.Request
	document fetch.Document
	err      error
}

func (s *stubFetcher) Fetch(_ context.Context, request fetch.Request) (fetch.Document, error) {
	s.request = request
	return s.document, s.err
}

type stubParser struct {
	document fetch.Document
	records  []parse.ParsedSearchRecord
	warnings []model.Warning
	err      error
}

func (s *stubParser) Parse(_ context.Context, _ model.SourceDescriptor, document fetch.Document) ([]parse.ParsedSearchRecord, []model.Warning, error) {
	s.document = document
	return s.records, s.warnings, s.err
}

type stubNormalizer struct {
	records    []parse.ParsedSearchRecord
	candidates []model.SearchCandidate
	warnings   []model.Warning
	err        error
}

func (s *stubNormalizer) Normalize(_ context.Context, _ model.SourceDescriptor, records []parse.ParsedSearchRecord) ([]model.SearchCandidate, []model.Warning, error) {
	s.records = records
	return s.candidates, s.warnings, s.err
}
