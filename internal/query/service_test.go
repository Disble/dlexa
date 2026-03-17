package query

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/source"
)

const (
	lookupErrFmt = "Lookup() error = %v"
	entryID1     = "entry-1"
	warningCode1 = "first-warning"
	problemCode1 = "first-problem"
	entryID2     = "entry-2"
	warningCode2 = "second-warning"
)

func TestLookupReturnsCachedResultAsCacheHit(t *testing.T) {
	request := model.LookupRequest{Query: "cache", Format: "json", Sources: []string{"demo"}}
	cached := model.LookupResult{Request: request, Entries: []model.Entry{{ID: "cached-entry"}}}
	store := &stubStore{getResult: cached, getOK: true}
	registry := &stubRegistry{}

	service := NewService(registry, store)
	result, err := service.Lookup(context.Background(), request)
	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}

	if !result.CacheHit {
		t.Fatal("Lookup() CacheHit = false, want true")
	}

	if !reflect.DeepEqual(result.Entries, cached.Entries) {
		t.Fatalf("Lookup() entries = %#v, want %#v", result.Entries, cached.Entries)
	}

	if registry.called {
		t.Fatal("SourcesFor() was called on cache hit")
	}

	if store.setCalls != 0 {
		t.Fatalf("Set() calls = %d, want 0", store.setCalls)
	}
}

func TestLookupAggregatesEntriesWarningsProblemsAndSourceResults(t *testing.T) {
	firstDescriptor := model.SourceDescriptor{Name: "demo", Priority: 1}
	secondDescriptor := model.SourceDescriptor{Name: "backup", Priority: 2}
	failingDescriptor := model.SourceDescriptor{Name: "broken", Priority: 3}

	registry := &stubRegistry{sources: []source.Source{
		&stubSource{
			descriptor: firstDescriptor,
			result: model.SourceResult{
				Source:   firstDescriptor,
				Entries:  []model.Entry{{ID: entryID1, Source: firstDescriptor.Name}},
				Warnings: []model.Warning{{Code: warningCode1, Source: firstDescriptor.Name}},
				Problems: []model.Problem{{Code: problemCode1, Source: firstDescriptor.Name, Severity: "warning"}},
			},
		},
		&stubSource{
			descriptor: secondDescriptor,
			result: model.SourceResult{
				Source:   secondDescriptor,
				Entries:  []model.Entry{{ID: entryID2, Source: secondDescriptor.Name}},
				Warnings: []model.Warning{{Code: warningCode2, Source: secondDescriptor.Name}},
			},
		},
		&stubSource{
			descriptor: failingDescriptor,
			err: model.NewProblemError(model.Problem{
				Code:     model.ProblemCodeDPDFetchFailed,
				Message:  "fetch failed",
				Source:   failingDescriptor.Name,
				Severity: model.ProblemSeverityError,
			}, errors.New("upstream timeout")),
		},
	}}

	service := NewService(registry, &stubStore{})
	request := model.LookupRequest{Query: "palabra", Format: "markdown", Sources: []string{"demo", "backup", "broken"}, NoCache: true}

	result, err := service.Lookup(context.Background(), request)
	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}

	if !reflect.DeepEqual(result.Request, request) {
		t.Fatalf("Lookup() request = %#v, want %#v", result.Request, request)
	}

	if len(result.Sources) != 2 {
		t.Fatalf("Lookup() sources = %d, want 2", len(result.Sources))
	}

	if gotIDs := []string{result.Entries[0].ID, result.Entries[1].ID}; !reflect.DeepEqual(gotIDs, []string{entryID1, entryID2}) {
		t.Fatalf("Lookup() entry IDs = %#v, want %#v", gotIDs, []string{entryID1, entryID2})
	}

	if gotCodes := []string{result.Warnings[0].Code, result.Warnings[1].Code}; !reflect.DeepEqual(gotCodes, []string{warningCode1, warningCode2}) {
		t.Fatalf("Lookup() warning codes = %#v, want %#v", gotCodes, []string{warningCode1, warningCode2})
	}

	if len(result.Problems) != 2 {
		t.Fatalf("Lookup() problems = %d, want 2", len(result.Problems))
	}

	if gotCodes := []string{result.Problems[0].Code, result.Problems[1].Code}; !reflect.DeepEqual(gotCodes, []string{problemCode1, model.ProblemCodeDPDFetchFailed}) {
		t.Fatalf("Lookup() problem codes = %#v, want %#v", gotCodes, []string{problemCode1, model.ProblemCodeDPDFetchFailed})
	}

	if result.Problems[1].Source != failingDescriptor.Name {
		t.Fatalf("Lookup() failing problem source = %q, want %q", result.Problems[1].Source, failingDescriptor.Name)
	}

	if result.GeneratedAt.IsZero() {
		t.Fatal("Lookup() GeneratedAt is zero")
	}
}

func TestLookupAggregatesStructuredMissesWithoutTurningThemIntoProblems(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "dpd", Priority: 1}
	miss := model.LookupMiss{
		Kind:   model.LookupMissKindRelatedEntry,
		Query:  "alicuota",
		Source: descriptor.Name,
		Suggestion: &model.LookupSuggestion{
			Kind:        "related_entry",
			DisplayText: "alícuota",
			EntryID:     "alícuota",
			URL:         "https://www.rae.es/dpd/alícuota",
		},
	}
	registry := &stubRegistry{sources: []source.Source{
		&stubSource{descriptor: descriptor, result: model.SourceResult{Source: descriptor, Miss: &miss}},
	}}

	service := NewService(registry, &stubStore{})
	result, err := service.Lookup(context.Background(), model.LookupRequest{Query: "alicuota", Sources: []string{"dpd"}, NoCache: true})
	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}
	if len(result.Misses) != 1 {
		t.Fatalf("Lookup() misses = %#v, want 1 structured miss", result.Misses)
	}
	if result.Misses[0].Kind != model.LookupMissKindRelatedEntry {
		t.Fatalf("miss kind = %q", result.Misses[0].Kind)
	}
	if len(result.Problems) != 0 {
		t.Fatalf("Lookup() problems = %#v, want none", result.Problems)
	}
}

func TestLookupFallsBackToGenericProblemForUntypedErrors(t *testing.T) {
	failingDescriptor := model.SourceDescriptor{Name: "broken"}
	registry := &stubRegistry{sources: []source.Source{
		&stubSource{descriptor: failingDescriptor, err: errors.New("boom")},
	}}

	service := NewService(registry, &stubStore{})
	result, err := service.Lookup(context.Background(), model.LookupRequest{Query: "palabra", Sources: []string{"broken"}, NoCache: true})
	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}

	if len(result.Problems) != 1 {
		t.Fatalf("Lookup() problems = %d, want 1", len(result.Problems))
	}

	want := model.Problem{
		Code:     model.ProblemCodeSourceLookupFailed,
		Message:  "boom",
		Source:   failingDescriptor.Name,
		Severity: model.ProblemSeverityError,
	}
	if !reflect.DeepEqual(result.Problems[0], want) {
		t.Fatalf("Lookup() problem = %#v, want %#v", result.Problems[0], want)
	}
}

type stubRegistry struct {
	sources []source.Source
	err     error
	called  bool
}

func (r *stubRegistry) SourcesFor(model.LookupRequest) ([]source.Source, error) {
	r.called = true
	if r.err != nil {
		return nil, r.err
	}
	return r.sources, nil
}

type stubStore struct {
	getResult model.LookupResult
	getOK     bool
	getErr    error
	setErr    error
	getCalls  int
	setCalls  int
}

func (s *stubStore) Get(_ context.Context, _ string) (model.LookupResult, bool, error) {
	s.getCalls++
	return s.getResult, s.getOK, s.getErr
}

func (s *stubStore) Set(_ context.Context, _ string, _ model.LookupResult) error {
	s.setCalls++
	return s.setErr
}

type stubSource struct {
	descriptor model.SourceDescriptor
	result     model.SourceResult
	err        error
}

func (s *stubSource) Descriptor() model.SourceDescriptor {
	return s.descriptor
}

func (s *stubSource) Lookup(context.Context, model.LookupRequest) (model.SourceResult, error) {
	return s.result, s.err
}
