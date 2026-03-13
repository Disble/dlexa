package query

import (
	"context"
	"time"

	"github.com/gentleman-programming/dlexa/internal/cache"
	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/source"
)

type LookupService struct {
	registry source.Registry
	cache    cache.Store
}

func NewService(registry source.Registry, store cache.Store) *LookupService {
	return &LookupService{registry: registry, cache: store}
}

func (s *LookupService) Lookup(ctx context.Context, request model.LookupRequest) (model.LookupResult, error) {
	cacheKey := cache.BuildKey(request)
	if s.cache != nil && !request.NoCache {
		cached, ok, err := s.cache.Get(ctx, cacheKey)
		if err != nil {
			return model.LookupResult{}, err
		}
		if ok {
			cached.CacheHit = true
			return cached, nil
		}
	}

	resolvedSources, err := s.registry.SourcesFor(request)
	if err != nil {
		return model.LookupResult{}, err
	}

	result := model.LookupResult{
		Request:     request,
		GeneratedAt: time.Now().UTC(),
	}

	for _, item := range resolvedSources {
		sourceResult, err := item.Lookup(ctx, request)
		if err != nil {
			result.Problems = append(result.Problems, model.Problem{
				Code:     "source_lookup_failed",
				Message:  err.Error(),
				Source:   item.Descriptor().Name,
				Severity: "error",
			})
			continue
		}

		result.Sources = append(result.Sources, sourceResult)
		result.Entries = append(result.Entries, sourceResult.Entries...)
		result.Warnings = append(result.Warnings, sourceResult.Warnings...)
		result.Problems = append(result.Problems, sourceResult.Problems...)
	}

	if s.cache != nil && !request.NoCache {
		if err := s.cache.Set(ctx, cacheKey, result); err != nil {
			return model.LookupResult{}, err
		}
	}

	return result, nil
}
