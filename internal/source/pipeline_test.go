package source

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
	parseengine "github.com/Disble/dlexa/internal/parse/engine"
)

const (
	parseWarningCode     = "parse-warning"
	normalizeWarningCode = "normalize-warning"
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
		warnings:       []model.Warning{{Code: parseWarningCode, Source: descriptor.Name}},
		expectedBody:   []byte("raw body"),
		expectedURL:    "https://example.invalid/demo/palabra",
		expectedSource: descriptor,
	}
	normalizer := &recordingNormalizer{
		calls:          &callOrder,
		result:         normalize.Result{Entries: normalizedEntries, Warnings: []model.Warning{{Code: normalizeWarningCode, Source: descriptor.Name}}},
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

	if gotCodes := []string{result.Warnings[0].Code, result.Warnings[1].Code}; !reflect.DeepEqual(gotCodes, []string{parseWarningCode, normalizeWarningCode}) {
		t.Fatalf("Lookup() warning codes = %#v, want %#v", gotCodes, []string{parseWarningCode, normalizeWarningCode})
	}

	if !result.FetchedAt.Equal(retrievedAt) {
		t.Fatalf("Lookup() fetchedAt = %v, want %v", result.FetchedAt, retrievedAt)
	}
}

func TestPipelineSourcePropagatesStructuredMissWithoutProblemFallback(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "dpd", DisplayName: "DPD"}
	callOrder := make([]string, 0, 3)
	parsedResult := parse.Result{Miss: &parse.ParsedLookupMiss{Kind: parse.ParsedLookupMissKindGenericNotFound, Query: "alicuota"}}
	normalizedMiss := &model.LookupMiss{
		Kind:   model.LookupMissKindGenericNotFound,
		Query:  "alicuota",
		Source: descriptor.Name,
		NextAction: &model.LookupNextAction{
			Kind:    model.LookupNextActionKindSearch,
			Query:   "alicuota",
			Command: "dlexa search alicuota",
		},
	}

	pipeline := NewPipelineSource(
		descriptor,
		&recordingFetcher{calls: &callOrder, document: fetch.Document{Body: []byte("body")}},
		&recordingParser{calls: &callOrder, result: parsedResult, expectedBody: []byte("body"), expectedSource: descriptor},
		&recordingNormalizer{calls: &callOrder, result: normalize.Result{Miss: normalizedMiss}, expectedResult: parsedResult, expectedSource: descriptor},
	)

	result, err := pipeline.Lookup(context.Background(), model.LookupRequest{Query: "alicuota"})
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if result.Miss == nil {
		t.Fatal("Lookup() miss = nil")
	}
	if *result.Miss != *normalizedMiss {
		t.Fatalf("Lookup() miss = %#v, want %#v", result.Miss, normalizedMiss)
	}
	if len(result.Problems) != 0 {
		t.Fatalf("Lookup() problems = %#v, want none", result.Problems)
	}
}

func TestNewEnginePipelineSourcePreservesLegacyParserBehavior(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "dpd", DisplayName: "DPD"}
	callOrder := make([]string, 0, 3)
	parsedResult := parse.Result{Articles: []parse.ParsedArticle{{Lemma: "Parsed"}}}
	legacyParser := &recordingParser{
		calls:          &callOrder,
		result:         parsedResult,
		warnings:       []model.Warning{{Code: parseWarningCode, Source: descriptor.Name}},
		expectedBody:   []byte("raw body"),
		expectedURL:    "https://example.invalid/demo/palabra",
		expectedSource: descriptor,
	}

	pipeline := NewEnginePipelineSource(
		descriptor,
		&recordingFetcher{calls: &callOrder, document: fetch.Document{URL: "https://example.invalid/demo/palabra", Body: []byte("raw body")}},
		parseengine.AdaptLegacyArticleParser(legacyParser),
		&recordingNormalizer{calls: &callOrder, result: normalize.Result{Entries: []model.Entry{{ID: "normalized", Source: descriptor.Name}}}, expectedResult: parsedResult, expectedSource: descriptor},
	)

	result, err := pipeline.Lookup(context.Background(), model.LookupRequest{Query: "palabra"})
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if !reflect.DeepEqual(callOrder, []string{"fetch", "parse", "normalize"}) {
		t.Fatalf("call order = %#v, want %#v", callOrder, []string{"fetch", "parse", "normalize"})
	}
	if len(result.Entries) != 1 || result.Entries[0].ID != "normalized" {
		t.Fatalf("Lookup() entries = %#v, want normalized result through engine adapter", result.Entries)
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
	result         normalize.Result
	expectedResult parse.Result
	expectedSource model.SourceDescriptor
	inputResult    parse.Result
	descriptor     model.SourceDescriptor
}

func (n *recordingNormalizer) Normalize(_ context.Context, descriptor model.SourceDescriptor, result parse.Result) (normalize.Result, error) {
	*n.calls = append(*n.calls, "normalize")
	n.descriptor = descriptor
	n.inputResult = result
	if !reflect.DeepEqual(descriptor, n.expectedSource) {
		return normalize.Result{}, &testError{message: "normalizer received unexpected source descriptor"}
	}
	if !reflect.DeepEqual(result, n.expectedResult) {
		return normalize.Result{}, &testError{message: "normalizer received unexpected entries"}
	}
	return n.result, nil
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
