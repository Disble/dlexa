package search

import (
	"context"
	"time"

	"github.com/Disble/dlexa/internal/cache"
	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// Searcher defines the contract for performing DPD entry searches.
type Searcher interface {
	Search(ctx context.Context, request model.SearchRequest) (model.SearchResult, error)
}

// Parser decodes fetched entry-search payloads into parsed records.
type Parser interface {
	Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]parse.ParsedSearchRecord, []model.Warning, error)
}

// Normalizer converts parsed entry-search records into normalized search candidates.
type Normalizer interface {
	Normalize(ctx context.Context, descriptor model.SourceDescriptor, records []parse.ParsedSearchRecord) ([]model.SearchCandidate, []model.Warning, error)
}

// Service orchestrates cache-aside DPD entry search execution.
type Service struct {
	descriptor model.SourceDescriptor
	fetcher    fetch.Fetcher
	parser     Parser
	normalizer Normalizer
	cache      cache.SearchStore
	now        func() time.Time
}

// NewService creates a search service backed by the given fetch/parse/normalize/cache adapters.
func NewService(descriptor model.SourceDescriptor, fetcher fetch.Fetcher, parser Parser, normalizer Normalizer, store cache.SearchStore) *Service {
	return &Service{descriptor: descriptor, fetcher: fetcher, parser: parser, normalizer: normalizer, cache: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) cacheResult(ctx context.Context, cacheKey string, request model.SearchRequest, result model.SearchResult) error {
	if s.cache == nil || request.NoCache {
		return nil
	}

	cacheCopy := result
	cacheCopy.Request = model.SearchRequest{Query: cache.NormalizeSearchQuery(request.Query)}

	return s.cache.Set(ctx, cacheKey, cacheCopy)
}

// Search runs a DPD entry search using a format-neutral cached normalized result when available.
func (s *Service) Search(ctx context.Context, request model.SearchRequest) (model.SearchResult, error) {
	cacheKey := cache.BuildSearchKey(request)
	if s.cache != nil && !request.NoCache {
		cached, ok, err := s.cache.Get(ctx, cacheKey)
		if err != nil {
			return model.SearchResult{}, err
		}
		if ok {
			cached.Request = request
			cached.CacheHit = true
			return cached, nil
		}
	}

	document, err := s.fetcher.Fetch(ctx, fetch.Request{Query: request.Query, Source: s.descriptor})
	if err != nil {
		return model.SearchResult{}, err
	}
	records, parseWarnings, err := s.parser.Parse(ctx, s.descriptor, document)
	if err != nil {
		return model.SearchResult{}, err
	}
	candidates, normalizeWarnings, err := s.normalizer.Normalize(ctx, s.descriptor, records)
	if err != nil {
		return model.SearchResult{}, err
	}

	result := model.SearchResult{
		Request:     request,
		Candidates:  candidates,
		Warnings:    append(append([]model.Warning(nil), parseWarnings...), normalizeWarnings...),
		GeneratedAt: s.now(),
	}

	if err := s.cacheResult(ctx, cacheKey, request, result); err != nil {
		return model.SearchResult{}, err
	}

	return result, nil
}
