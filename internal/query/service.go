package query

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Disble/dlexa/internal/cache"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/source"
)

// sourceOutcome captures the result of a single source lookup goroutine.
type sourceOutcome struct {
	source source.Source
	result model.SourceResult
	err    error
}

// LookupService orchestrates parallel source lookups with caching.
type LookupService struct {
	registry source.Registry
	cache    cache.Store
}

// NewService creates a LookupService backed by the given registry and cache.
func NewService(registry source.Registry, store cache.Store) *LookupService {
	return &LookupService{registry: registry, cache: store}
}

// Lookup fans out to all matching sources in parallel and merges results.
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

	// Fan-out: launch one goroutine per source.
	outcomes := make(chan sourceOutcome, len(resolvedSources))
	var wg sync.WaitGroup

	for _, item := range resolvedSources {
		wg.Add(1)
		go func(src source.Source) {
			defer wg.Done()
			srcResult, srcErr := src.Lookup(ctx, request)
			outcomes <- sourceOutcome{source: src, result: srcResult, err: srcErr}
		}(item)
	}

	// Close channel after all goroutines complete.
	go func() {
		wg.Wait()
		close(outcomes)
	}()

	// Fan-in: collect all outcomes from the channel.
	var collected []sourceOutcome
	for outcome := range outcomes {
		collected = append(collected, outcome)
	}

	// Sort outcomes by source priority (ascending) for deterministic ordering.
	sort.SliceStable(collected, func(i, j int) bool {
		return collected[i].source.Descriptor().Priority < collected[j].source.Descriptor().Priority
	})

	// Aggregate results in priority order.
	for _, outcome := range collected {
		if outcome.err != nil {
			problem, ok := model.AsProblem(outcome.err)
			if !ok {
				problem = model.Problem{
					Code:     model.ProblemCodeSourceLookupFailed,
					Message:  outcome.err.Error(),
					Source:   outcome.source.Descriptor().Name,
					Severity: model.ProblemSeverityError,
				}
			} else if problem.Source == "" {
				problem.Source = outcome.source.Descriptor().Name
			}

			result.Problems = append(result.Problems, problem)
			continue
		}

		result.Sources = append(result.Sources, outcome.result)
		result.Entries = append(result.Entries, outcome.result.Entries...)
		result.Warnings = append(result.Warnings, outcome.result.Warnings...)
		result.Problems = append(result.Problems, outcome.result.Problems...)
	}

	if s.cache != nil && !request.NoCache {
		if err := s.cache.Set(ctx, cacheKey, result); err != nil {
			return model.LookupResult{}, err
		}
	}

	return result, nil
}
