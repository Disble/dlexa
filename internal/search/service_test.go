package search

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/Disble/dlexa/internal/cache"
	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

const searchErrFormat = "Search() error = %v"

func TestServiceReturnsCachedResultAndRefreshesRequestFields(t *testing.T) {
	request := model.SearchRequest{Query: " Abu Dhabi ", Format: "json"}
	stored := model.SearchResult{
		Request:    model.SearchRequest{Query: "Abu Dhabi", Format: "markdown", Sources: []string{"search"}},
		Candidates: []model.SearchCandidate{{Title: "Abu Dabi"}},
	}
	store := &stubSearchStore{getResult: stored, getOK: true}
	provider := &providerStub{descriptor: model.SourceDescriptor{Name: "search", Priority: 1}}
	service := NewService(NewStaticRegistry("search", provider), store, 1, "search")

	result, err := service.Search(context.Background(), request)
	if err != nil {
		t.Fatalf(searchErrFormat, err)
	}
	if !result.CacheHit {
		t.Fatal("CacheHit = false, want true")
	}
	if provider.calls.Load() != 0 {
		t.Fatalf("provider calls = %d, want 0 on cache hit", provider.calls.Load())
	}
	if !reflect.DeepEqual(result.Request, model.SearchRequest{Query: " Abu Dhabi ", Format: "json", Sources: []string{"search"}}) {
		t.Fatalf("Request = %#v", result.Request)
	}
}

func TestServiceReadsLegacyCacheForDefaultProviderAndRewritesV2Key(t *testing.T) {
	legacyResult := model.SearchResult{Request: model.SearchRequest{Query: "tilde"}, Candidates: []model.SearchCandidate{{Title: "solo"}}}
	store := &stubSearchStore{entries: map[string]model.SearchResult{cache.BuildLegacySearchKey(model.SearchRequest{Query: "tilde"}): legacyResult}}
	provider := &providerStub{descriptor: model.SourceDescriptor{Name: "search", Priority: 1}}
	service := NewService(NewStaticRegistry("search", provider), store, 1, "search")

	result, err := service.Search(context.Background(), model.SearchRequest{Query: "tilde"})
	if err != nil {
		t.Fatalf(searchErrFormat, err)
	}
	if !result.CacheHit {
		t.Fatal("CacheHit = false, want true")
	}
	if provider.calls.Load() != 0 {
		t.Fatalf("provider calls = %d, want 0 when legacy cache hits", provider.calls.Load())
	}
	if _, ok := store.entries[cache.BuildSearchKey(model.SearchRequest{Query: "tilde", Sources: []string{"search"}})]; !ok {
		t.Fatal("expected legacy cache hit to rewrite v2 cache key")
	}
}

func TestServiceDegradesWhenCacheReadFails(t *testing.T) {
	provider := &providerStub{descriptor: model.SourceDescriptor{Name: "search", Priority: 1}, result: model.SearchResult{Candidates: []model.SearchCandidate{{Title: "bien"}}}}
	store := &stubSearchStore{getErr: errors.New("cache unavailable")}
	service := NewService(NewStaticRegistry("search", provider), store, 1, "search")

	result, err := service.Search(context.Background(), model.SearchRequest{Query: "tilde", Format: "json"})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("Candidates len = %d, want 1", len(result.Candidates))
	}
	if store.setCalls != 1 {
		t.Fatalf("Set() calls = %d, want 1 after degraded cache read", store.setCalls)
	}
}

func TestServiceDegradesWhenCacheWriteFails(t *testing.T) {
	provider := &providerStub{descriptor: model.SourceDescriptor{Name: "search", Priority: 1}, result: model.SearchResult{Candidates: []model.SearchCandidate{{Title: "bien"}}}}
	service := NewService(NewStaticRegistry("search", provider), &stubSearchStore{setErr: errors.New("cache unavailable")}, 1, "search")

	result, err := service.Search(context.Background(), model.SearchRequest{Query: "tilde"})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("Candidates len = %d, want 1", len(result.Candidates))
	}
}

func TestServiceCoalescesConcurrentCacheMissesPerProvider(t *testing.T) {
	provider := &providerStub{
		descriptor: model.SourceDescriptor{Name: "search", Priority: 1},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: "bien"}}},
		started:    make(chan struct{}, 2),
		release:    make(chan struct{}),
	}
	service := NewService(NewStaticRegistry("search", provider), &stubSearchStore{}, 1, "search")

	type outcome struct {
		result model.SearchResult
		err    error
	}
	results := make(chan outcome, 2)
	requestA := model.SearchRequest{Query: " Abu   Dhabi ", Format: "json"}
	requestB := model.SearchRequest{Query: "Abu Dhabi", Format: "markdown"}

	go func() {
		result, err := service.Search(context.Background(), requestA)
		results <- outcome{result: result, err: err}
	}()
	provider.waitStarted(t)
	go func() {
		result, err := service.Search(context.Background(), requestB)
		results <- outcome{result: result, err: err}
	}()
	time.Sleep(25 * time.Millisecond)
	close(provider.release)

	first := <-results
	second := <-results
	if first.err != nil {
		t.Fatalf(searchErrFormat, first.err)
	}
	if second.err != nil {
		t.Fatalf(searchErrFormat, second.err)
	}
	if got := provider.calls.Load(); got != 1 {
		t.Fatalf("provider calls = %d, want 1", got)
	}
	wantA := model.SearchRequest{Query: requestA.Query, Format: requestA.Format, Sources: []string{"search"}}
	wantB := model.SearchRequest{Query: requestB.Query, Format: requestB.Format, Sources: []string{"search"}}
	if !reflect.DeepEqual(first.result.Request, wantA) && !reflect.DeepEqual(second.result.Request, wantA) {
		t.Fatalf("expected one requestA result, got %#v and %#v", first.result.Request, second.result.Request)
	}
	if !reflect.DeepEqual(first.result.Request, wantB) && !reflect.DeepEqual(second.result.Request, wantB) {
		t.Fatalf("expected one requestB result, got %#v and %#v", first.result.Request, second.result.Request)
	}
}

func TestServiceNoCacheBypassesCoalescing(t *testing.T) {
	provider := &providerStub{
		descriptor: model.SourceDescriptor{Name: "search", Priority: 1},
		result:     model.SearchResult{Candidates: []model.SearchCandidate{{Title: "bien"}}},
		started:    make(chan struct{}, 2),
		release:    make(chan struct{}),
	}
	service := NewService(NewStaticRegistry("search", provider), &stubSearchStore{}, 1, "search")

	results := make(chan error, 2)
	request := model.SearchRequest{Query: "Abu Dhabi", NoCache: true}
	for range 2 {
		go func() {
			_, err := service.Search(context.Background(), request)
			results <- err
		}()
	}
	provider.waitStarted(t)
	provider.waitStarted(t)
	close(provider.release)

	for range 2 {
		if err := <-results; err != nil {
			t.Fatalf(searchErrFormat, err)
		}
	}
	if got := provider.calls.Load(); got != 2 {
		t.Fatalf("provider calls = %d, want 2 when NoCache=true", got)
	}
}

type stubSearchStore struct {
	entries   map[string]model.SearchResult
	getResult model.SearchResult
	getOK     bool
	getErr    error
	setErr    error
	mu        sync.Mutex
	getCalls  int
	setCalls  int
	setResult model.SearchResult
	setKey    string
}

func (s *stubSearchStore) Get(_ context.Context, key string) (model.SearchResult, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getCalls++
	if s.getErr != nil {
		return model.SearchResult{}, false, s.getErr
	}
	if s.entries != nil {
		result, ok := s.entries[key]
		return result, ok, nil
	}
	return s.getResult, s.getOK, nil
}

func (s *stubSearchStore) Set(_ context.Context, key string, result model.SearchResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setCalls++
	s.setKey = key
	s.setResult = result
	if s.entries == nil {
		s.entries = make(map[string]model.SearchResult)
	}
	s.entries[key] = result
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
