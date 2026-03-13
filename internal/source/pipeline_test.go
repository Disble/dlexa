package source

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/gentleman-programming/dlexa/internal/fetch"
	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/parse"
)

func TestPipelineSourceFetchesParsesAndNormalizesInOrder(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "demo", DisplayName: "Demo"}
	retrievedAt := time.Date(2026, time.March, 13, 15, 0, 0, 0, time.UTC)
	callOrder := make([]string, 0, 3)
	parsedResult := parse.Result{Articles: []parse.ParsedArticle{{Lemma: "Parsed"}}}
	normalizedEntries := []model.Entry{{ID: "normalized", Headword: "Normalized", Source: descriptor.Name}}

	fetcher := &recordingFetcher{
		calls: &callOrder,
		document: fetch.Document{
			URL:         "https://example.invalid/demo/palabra",
			ContentType: "text/markdown",
			Body:        []byte("raw body"),
			RetrievedAt: retrievedAt,
		},
	}
	parser := &recordingParser{
		calls:          &callOrder,
		result:         parsedResult,
		warnings:       []model.Warning{{Code: "parse-warning", Source: descriptor.Name}},
		expectedBody:   []byte("raw body"),
		expectedURL:    "https://example.invalid/demo/palabra",
		expectedSource: descriptor,
	}
	normalizer := &recordingNormalizer{
		calls:          &callOrder,
		entries:        normalizedEntries,
		warnings:       []model.Warning{{Code: "normalize-warning", Source: descriptor.Name}},
		expectedResult: parsedResult,
		expectedSource: descriptor,
	}

	pipeline := NewPipelineSource(descriptor, fetcher, parser, normalizer)
	result, err := pipeline.Lookup(context.Background(), model.LookupRequest{Query: "palabra"})
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}

	if !reflect.DeepEqual(callOrder, []string{"fetch", "parse", "normalize"}) {
		t.Fatalf("call order = %#v, want %#v", callOrder, []string{"fetch", "parse", "normalize"})
	}

	if !reflect.DeepEqual(result.Entries, normalizedEntries) {
		t.Fatalf("Lookup() entries = %#v, want %#v", result.Entries, normalizedEntries)
	}

	if gotCodes := []string{result.Warnings[0].Code, result.Warnings[1].Code}; !reflect.DeepEqual(gotCodes, []string{"parse-warning", "normalize-warning"}) {
		t.Fatalf("Lookup() warning codes = %#v, want %#v", gotCodes, []string{"parse-warning", "normalize-warning"})
	}

	if !result.FetchedAt.Equal(retrievedAt) {
		t.Fatalf("Lookup() fetchedAt = %v, want %v", result.FetchedAt, retrievedAt)
	}
}

type recordingFetcher struct {
	calls    *[]string
	document fetch.Document
	request  fetch.Request
}

func (f *recordingFetcher) Fetch(_ context.Context, request fetch.Request) (fetch.Document, error) {
	*f.calls = append(*f.calls, "fetch")
	f.request = request
	return f.document, nil
}

type recordingParser struct {
	calls          *[]string
	result         parse.Result
	warnings       []model.Warning
	expectedBody   []byte
	expectedURL    string
	expectedSource model.SourceDescriptor
	document       fetch.Document
	descriptor     model.SourceDescriptor
}

func (p *recordingParser) Parse(_ context.Context, descriptor model.SourceDescriptor, document fetch.Document) (parse.Result, []model.Warning, error) {
	*p.calls = append(*p.calls, "parse")
	p.descriptor = descriptor
	p.document = document
	if !reflect.DeepEqual(descriptor, p.expectedSource) {
		return parse.Result{}, nil, &testError{message: "parser received unexpected source descriptor"}
	}
	if !reflect.DeepEqual(document.Body, p.expectedBody) || document.URL != p.expectedURL {
		return parse.Result{}, nil, &testError{message: "parser received unexpected document"}
	}
	return p.result, p.warnings, nil
}

type recordingNormalizer struct {
	calls          *[]string
	entries        []model.Entry
	warnings       []model.Warning
	expectedResult parse.Result
	expectedSource model.SourceDescriptor
	inputResult    parse.Result
	descriptor     model.SourceDescriptor
}

func (n *recordingNormalizer) Normalize(_ context.Context, descriptor model.SourceDescriptor, result parse.Result) ([]model.Entry, []model.Warning, error) {
	*n.calls = append(*n.calls, "normalize")
	n.descriptor = descriptor
	n.inputResult = result
	if !reflect.DeepEqual(descriptor, n.expectedSource) {
		return nil, nil, &testError{message: "normalizer received unexpected source descriptor"}
	}
	if !reflect.DeepEqual(result, n.expectedResult) {
		return nil, nil, &testError{message: "normalizer received unexpected entries"}
	}
	return n.entries, n.warnings, nil
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
