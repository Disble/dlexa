package search

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/cache"
	"github.com/Disble/dlexa/internal/model"
)

func TestServiceUsesRequestedSourcesAndOrdersCandidatesByPriority(t *testing.T) {
	providerP3 := &providerStub{
		descriptor: model.SourceDescriptor{Name: "priority-3", Priority: 3},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: "p3"}}},
	}
	providerP1 := &providerStub{
		descriptor: model.SourceDescriptor{Name: "priority-1", Priority: 1},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: "p1"}}},
	}
	providerP2 := &providerStub{
		descriptor: model.SourceDescriptor{Name: "priority-2", Priority: 2},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: "p2"}}},
	}

	service := NewService(NewStaticRegistry("priority-1", providerP3, providerP1, providerP2), cache.NewSearchMemoryStore(), 2, "priority-1")
	request := model.SearchRequest{Query: "ordenado", Sources: []string{"priority-3", "priority-1"}, NoCache: true}

	result, err := service.Search(context.Background(), request)
	if err != nil {
		t.Fatalf(searchErrFormat, err)
	}

	if providerP2.calls.Load() != 0 {
		t.Fatalf("providerP2 calls = %d, want 0 for unrequested source", providerP2.calls.Load())
	}
	if got := candidateTitles(result.Candidates); !reflect.DeepEqual(got, []string{"p1", "p3"}) {
		t.Fatalf("candidate order = %v, want [p1 p3]", got)
	}
	assertSingletonSourceRequest(t, providerP1, "priority-1")
	assertSingletonSourceRequest(t, providerP3, "priority-3")
}

func TestServiceBoundsConcurrentProviderSearches(t *testing.T) {
	tracker := &concurrencyTracker{}
	providerA := newBlockingProvider("a", 1, tracker)
	providerB := newBlockingProvider("b", 2, tracker)
	providerC := newBlockingProvider("c", 3, tracker)
	service := NewService(NewStaticRegistry("a", providerA, providerB, providerC), cache.NewSearchMemoryStore(), 2, "a")

	resultCh := make(chan error, 1)
	go func() {
		_, err := service.Search(context.Background(), model.SearchRequest{Query: "límite", Sources: []string{"a", "b", "c"}, NoCache: true})
		resultCh <- err
	}()

	waitForProviderCalls(t, []*providerStub{providerA, providerB, providerC}, 2)
	if got := totalProviderCalls(providerA, providerB, providerC); got != 2 {
		t.Fatalf("provider calls before release = %d, want 2", got)
	}
	time.Sleep(50 * time.Millisecond)
	if got := totalProviderCalls(providerA, providerB, providerC); got != 2 {
		t.Fatalf("provider calls while blocked = %d, want 2", got)
	}

	close(providerA.release)
	close(providerB.release)
	close(providerC.release)

	if err := <-resultCh; err != nil {
		t.Fatalf(searchErrFormat, err)
	}
	if got := tracker.maxConcurrentSeen(); got != 2 {
		t.Fatalf("max concurrent providers = %d, want 2", got)
	}
}

func TestServiceReturnsPartialCacheHitsAndFreshResultsTogether(t *testing.T) {
	store := cache.NewSearchMemoryStore()
	cachedProvider := &providerStub{
		descriptor: model.SourceDescriptor{Name: "search", Priority: 1},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: "cached"}}},
	}
	freshProvider := &providerStub{
		descriptor: model.SourceDescriptor{Name: "academia", Priority: 2},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: "fresh"}}},
	}
	cachedKey := cache.BuildSearchKey(model.SearchRequest{Query: "tilde", Sources: []string{"search"}})
	if err := store.Set(context.Background(), cachedKey, model.SearchResult{Candidates: []model.SearchCandidate{{Title: "cached"}}}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	service := NewService(NewStaticRegistry("search", cachedProvider, freshProvider), store, 2, "search")
	result, err := service.Search(context.Background(), model.SearchRequest{Query: "tilde", Sources: []string{"search", "academia"}})
	if err != nil {
		t.Fatalf(searchErrFormat, err)
	}
	if result.CacheHit {
		t.Fatal("CacheHit = true, want false for partial cache hit")
	}
	if got := candidateTitles(result.Candidates); !reflect.DeepEqual(got, []string{"cached", "fresh"}) {
		t.Fatalf("candidate order = %v, want [cached fresh]", got)
	}
	if cachedProvider.calls.Load() != 0 {
		t.Fatalf("cached provider calls = %d, want 0", cachedProvider.calls.Load())
	}
	if freshProvider.calls.Load() != 1 {
		t.Fatalf("fresh provider calls = %d, want 1", freshProvider.calls.Load())
	}
}

func TestServiceReturnsProblemsWhenOneProviderFails(t *testing.T) {
	providerA := &providerStub{
		descriptor: model.SourceDescriptor{Name: "search", Priority: 1},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: "bien"}}},
	}
	providerB := &providerStub{
		descriptor: model.SourceDescriptor{Name: "academia", Priority: 2},
		err:        model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "academia unavailable", Severity: model.ProblemSeverityError}, errors.New("timeout")),
	}

	service := NewService(NewStaticRegistry("search", providerA, providerB), cache.NewSearchMemoryStore(), 2, "search")
	result, err := service.Search(context.Background(), model.SearchRequest{Query: "tilde", Sources: []string{"search", "academia"}, NoCache: true})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if got := candidateTitles(result.Candidates); !reflect.DeepEqual(got, []string{"bien"}) {
		t.Fatalf("candidate titles = %v, want [bien]", got)
	}
	if len(result.Problems) != 1 {
		t.Fatalf("Problems len = %d, want 1", len(result.Problems))
	}
	if result.Problems[0].Source != "academia" {
		t.Fatalf("Problems[0].Source = %q, want academia", result.Problems[0].Source)
	}
}

func TestServiceReturnsTopLevelErrorWhenAllProvidersFail(t *testing.T) {
	providerA := &providerStub{
		descriptor: model.SourceDescriptor{Name: "search", Priority: 1},
		err:        model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "search unavailable", Source: "search", Severity: model.ProblemSeverityError}, errors.New("timeout")),
	}
	providerB := &providerStub{
		descriptor: model.SourceDescriptor{Name: "academia", Priority: 2},
		err:        model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "academia unavailable", Source: "academia", Severity: model.ProblemSeverityError}, errors.New("timeout")),
	}

	service := NewService(NewStaticRegistry("search", providerA, providerB), cache.NewSearchMemoryStore(), 2, "search")
	if _, err := service.Search(context.Background(), model.SearchRequest{Query: "tilde", Sources: []string{"search", "academia"}, NoCache: true}); err == nil {
		t.Fatal("Search() error = nil, want top-level error when all providers fail")
	}
}

func candidateTitles(candidates []model.SearchCandidate) []string {
	titles := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		titles = append(titles, candidate.Title)
	}
	return titles
}

func assertSingletonSourceRequest(t *testing.T, provider *providerStub, want string) {
	t.Helper()
	requests := provider.recordedRequests()
	if len(requests) != 1 {
		t.Fatalf("provider %q requests = %d, want 1", provider.descriptor.Name, len(requests))
	}
	if got := requests[0].Sources; !reflect.DeepEqual(got, []string{want}) {
		t.Fatalf("provider %q request sources = %#v, want [%q]", provider.descriptor.Name, got, want)
	}
}

func newBlockingProvider(name string, priority int, tracker *concurrencyTracker) *providerStub {
	return &providerStub{
		descriptor: model.SourceDescriptor{Name: name, Priority: priority},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: name}}},
		started:    make(chan struct{}, 1),
		release:    make(chan struct{}),
		tracker:    tracker,
	}
}

func (p *providerStub) waitStarted(t *testing.T) {
	t.Helper()
	select {
	case <-p.started:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for provider %q start", p.descriptor.Name)
	}
}

func (p *providerStub) assertNotStartedYet(t *testing.T) {
	t.Helper()
	select {
	case <-p.started:
		t.Fatalf("provider %q started before concurrency slot freed", p.descriptor.Name)
	case <-time.After(50 * time.Millisecond):
	}
}

func totalProviderCalls(providers ...*providerStub) int {
	total := 0
	for _, provider := range providers {
		total += int(provider.calls.Load())
	}
	return total
}

func waitForProviderCalls(t *testing.T, providers []*providerStub, want int) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if totalProviderCalls(providers...) >= want {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %d provider calls; got %d", want, totalProviderCalls(providers...))
		case <-ticker.C:
		}
	}
}

type providerStub struct {
	descriptor model.SourceDescriptor
	result     model.SearchResult
	err        error
	started    chan struct{}
	release    chan struct{}
	tracker    *concurrencyTracker

	mu       sync.Mutex
	requests []model.SearchRequest

	calls atomic.Int32
}

func (p *providerStub) Descriptor() model.SourceDescriptor { return p.descriptor }

func (p *providerStub) Search(ctx context.Context, request model.SearchRequest) (model.SearchResult, error) {
	p.calls.Add(1)
	p.mu.Lock()
	p.requests = append(p.requests, request)
	p.mu.Unlock()

	if p.tracker != nil {
		p.tracker.enter()
		defer p.tracker.leave()
	}

	if p.started != nil {
		p.started <- struct{}{}
	}
	if p.release != nil {
		select {
		case <-p.release:
		case <-ctx.Done():
			return model.SearchResult{}, ctx.Err()
		}
	}
	if p.err != nil {
		return model.SearchResult{}, p.err
	}
	result := p.result
	result.Request = request
	if result.GeneratedAt.IsZero() {
		result.GeneratedAt = time.Date(2026, time.April, 8, 0, 0, 0, 0, time.UTC)
	}
	return result, nil
}

func (p *providerStub) recordedRequests() []model.SearchRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]model.SearchRequest(nil), p.requests...)
}

type concurrencyTracker struct {
	inFlight atomic.Int32
	maxSeen  atomic.Int32
}

func (t *concurrencyTracker) enter() {
	current := t.inFlight.Add(1)
	for {
		maxSeen := t.maxSeen.Load()
		if current <= maxSeen || t.maxSeen.CompareAndSwap(maxSeen, current) {
			return
		}
	}
}

func (t *concurrencyTracker) leave() {
	t.inFlight.Add(-1)
}

func (t *concurrencyTracker) maxConcurrentSeen() int {
	return int(t.maxSeen.Load())
}
