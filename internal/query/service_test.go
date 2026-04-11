package query

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	freshEntryID = "fresh-entry"
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

func TestLookupDegradesWhenCacheReadFails(t *testing.T) {
	request := model.LookupRequest{Query: "palabra", Format: "json", Sources: []string{"demo"}}
	store := &stubStore{getErr: errors.New("cache unavailable")}
	registry := &stubRegistry{sources: []source.Source{
		&stubSource{
			descriptor: model.SourceDescriptor{Name: "demo", Priority: 1},
			result: model.SourceResult{
				Source:  model.SourceDescriptor{Name: "demo", Priority: 1},
				Entries: []model.Entry{{ID: freshEntryID, Source: "demo"}},
			},
		},
	}}

	result, err := NewService(registry, store).Lookup(context.Background(), request)
	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}
	if len(result.Entries) != 1 || result.Entries[0].ID != freshEntryID {
		t.Fatalf("Lookup() entries = %#v, want fresh origin result", result.Entries)
	}
	if store.setCalls != 1 {
		t.Fatalf("Set() calls = %d, want 1 after degraded cache read", store.setCalls)
	}
}

func TestLookupDegradesWhenCacheWriteFails(t *testing.T) {
	request := model.LookupRequest{Query: "palabra", Format: "json", Sources: []string{"demo"}}
	store := &stubStore{setErr: errors.New("cache unavailable")}
	registry := &stubRegistry{sources: []source.Source{
		&stubSource{
			descriptor: model.SourceDescriptor{Name: "demo", Priority: 1},
			result: model.SourceResult{
				Source:  model.SourceDescriptor{Name: "demo", Priority: 1},
				Entries: []model.Entry{{ID: freshEntryID, Source: "demo"}},
			},
		},
	}}

	result, err := NewService(registry, store).Lookup(context.Background(), request)
	if err != nil {
		t.Fatalf(lookupErrFmt, err)
	}
	if len(result.Entries) != 1 || result.Entries[0].ID != freshEntryID {
		t.Fatalf("Lookup() entries = %#v, want fresh origin result", result.Entries)
	}
	if result.CacheHit {
		t.Fatal("Lookup() CacheHit = true, want false for degraded cache write")
	}
}

func TestLookupCoalescesConcurrentCacheMisses(t *testing.T) {
	request := model.LookupRequest{Query: "palabra", Format: "json", Sources: []string{"demo"}}
	lookupSource := newGatedLookupSource(model.SourceDescriptor{Name: "demo", Priority: 1}, model.SourceResult{
		Source:  model.SourceDescriptor{Name: "demo", Priority: 1},
		Entries: []model.Entry{{ID: freshEntryID, Source: "demo"}},
	})
	registry := &countingRegistry{sources: []source.Source{lookupSource}}
	service := NewService(registry, &stubStore{})

	type lookupOutcome struct {
		result model.LookupResult
		err    error
	}
	results := make(chan lookupOutcome, 2)

	go func() {
		result, err := service.Lookup(context.Background(), request)
		results <- lookupOutcome{result: result, err: err}
	}()
	lookupSource.waitStarted(t, 1)
	go func() {
		result, err := service.Lookup(context.Background(), request)
		results <- lookupOutcome{result: result, err: err}
	}()
	time.Sleep(25 * time.Millisecond)
	close(lookupSource.release)

	for i := 0; i < 2; i++ {
		outcome := <-results
		if outcome.err != nil {
			t.Fatalf(lookupErrFmt, outcome.err)
		}
		if len(outcome.result.Entries) != 1 || outcome.result.Entries[0].ID != freshEntryID {
			t.Fatalf("Lookup() entries = %#v, want coalesced fresh result", outcome.result.Entries)
		}
	}
	if got := lookupSource.calls.Load(); got != 1 {
		t.Fatalf("source Lookup() calls = %d, want 1", got)
	}
	if got := registry.calls.Load(); got != 1 {
		t.Fatalf("SourcesFor() calls = %d, want 1", got)
	}
}

func TestLookupNoCacheBypassesCoalescing(t *testing.T) {
	request := model.LookupRequest{Query: "palabra", Format: "json", Sources: []string{"demo"}, NoCache: true}
	lookupSource := newGatedLookupSource(model.SourceDescriptor{Name: "demo", Priority: 1}, model.SourceResult{
		Source:  model.SourceDescriptor{Name: "demo", Priority: 1},
		Entries: []model.Entry{{ID: freshEntryID, Source: "demo"}},
	})
	registry := &countingRegistry{sources: []source.Source{lookupSource}}
	service := NewService(registry, &stubStore{})

	results := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := service.Lookup(context.Background(), request)
			results <- err
		}()
	}
	lookupSource.waitStarted(t, 2)
	close(lookupSource.release)

	for i := 0; i < 2; i++ {
		if err := <-results; err != nil {
			t.Fatalf(lookupErrFmt, err)
		}
	}
	if got := lookupSource.calls.Load(); got != 2 {
		t.Fatalf("source Lookup() calls = %d, want 2 when NoCache=true", got)
	}
	if got := registry.calls.Load(); got != 2 {
		t.Fatalf("SourcesFor() calls = %d, want 2 when NoCache=true", got)
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
	mu        sync.Mutex
	getCalls  int
	setCalls  int
}

func (s *stubStore) Get(_ context.Context, _ string) (model.LookupResult, bool, error) {
	s.mu.Lock()
	s.getCalls++
	s.mu.Unlock()
	return s.getResult, s.getOK, s.getErr
}

func (s *stubStore) Set(_ context.Context, _ string, _ model.LookupResult) error {
	s.mu.Lock()
	s.setCalls++
	s.mu.Unlock()
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

type countingRegistry struct {
	sources []source.Source
	err     error
	calls   atomic.Int32
}

func (r *countingRegistry) SourcesFor(model.LookupRequest) ([]source.Source, error) {
	r.calls.Add(1)
	if r.err != nil {
		return nil, r.err
	}
	return r.sources, nil
}

type gatedLookupSource struct {
	descriptor model.SourceDescriptor
	result     model.SourceResult
	err        error
	started    chan struct{}
	release    chan struct{}
	calls      atomic.Int32
}

func newGatedLookupSource(descriptor model.SourceDescriptor, result model.SourceResult) *gatedLookupSource {
	return &gatedLookupSource{
		descriptor: descriptor,
		result:     result,
		started:    make(chan struct{}, 4),
		release:    make(chan struct{}),
	}
}

func (s *gatedLookupSource) Descriptor() model.SourceDescriptor {
	return s.descriptor
}

func (s *gatedLookupSource) Lookup(ctx context.Context, _ model.LookupRequest) (model.SourceResult, error) {
	s.calls.Add(1)
	s.started <- struct{}{}
	select {
	case <-s.release:
		return s.result, s.err
	case <-ctx.Done():
		return model.SourceResult{}, ctx.Err()
	}
}

func (s *gatedLookupSource) waitStarted(t *testing.T, want int) {
	t.Helper()
	for i := 0; i < want; i++ {
		select {
		case <-s.started:
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for source start %d/%d", i+1, want)
		}
	}
}
