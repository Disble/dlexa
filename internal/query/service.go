package query

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Disble/dlexa/internal/cache"
	"github.com/Disble/dlexa/internal/inflight"
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
	registry source.SourcesForer
	cache    cache.Store
	flights  inflight.Group[model.LookupResult]
}

// NewService creates a LookupService backed by the given registry and cache.
func NewService(registry source.SourcesForer, store cache.Store) *LookupService {
	return &LookupService{registry: registry, cache: store}
}

// RegistryForTesting exposes the wired lookup source registry for app wiring tests.
func (s *LookupService) RegistryForTesting() source.SourcesForer { return s.registry }

func (s *LookupService) lookupCachedResult(ctx context.Context, cacheKey string, request model.LookupRequest) (model.LookupResult, bool, error) {
	if s.cache == nil || request.NoCache {
		return model.LookupResult{}, false, nil
	}

	cached, ok, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		return model.LookupResult{}, false, nil
	}
	if !ok {
		return model.LookupResult{}, false, nil
	}

	cached.CacheHit = true
	return cached, true, nil
}

func collectOutcomes(ctx context.Context, request model.LookupRequest, resolvedSources []source.Source) []sourceOutcome {
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

	go func() {
		wg.Wait()
		close(outcomes)
	}()

	collected := make([]sourceOutcome, 0, len(resolvedSources))
	for outcome := range outcomes {
		collected = append(collected, outcome)
	}

	return collected
}

func sortOutcomesByPriority(collected []sourceOutcome) {
	sort.SliceStable(collected, func(i, j int) bool {
		return collected[i].source.Descriptor().Priority < collected[j].source.Descriptor().Priority
	})
}

func problemFromOutcome(outcome sourceOutcome) model.Problem {
	problem, ok := model.AsProblem(outcome.err)
	if !ok {
		return model.Problem{
			Code:     model.ProblemCodeSourceLookupFailed,
			Message:  outcome.err.Error(),
			Source:   outcome.source.Descriptor().Name,
			Severity: model.ProblemSeverityError,
		}
	}
	if problem.Source == "" {
		problem.Source = outcome.source.Descriptor().Name
	}

	return problem
}

func appendOutcome(result *model.LookupResult, outcome sourceOutcome) {
	if outcome.err != nil {
		result.Problems = append(result.Problems, problemFromOutcome(outcome))
		return
	}

	result.Sources = append(result.Sources, outcome.result)
	result.Entries = append(result.Entries, outcome.result.Entries...)
	if outcome.result.Miss != nil {
		result.Misses = append(result.Misses, *outcome.result.Miss)
	}
	result.Warnings = append(result.Warnings, outcome.result.Warnings...)
	result.Problems = append(result.Problems, outcome.result.Problems...)
}

func (s *LookupService) cacheResult(ctx context.Context, cacheKey string, request model.LookupRequest, result model.LookupResult) error {
	if s.cache == nil || request.NoCache {
		return nil
	}

	return s.cache.Set(ctx, cacheKey, result)
}

func (s *LookupService) lookupFreshResult(ctx context.Context, cacheKey string, request model.LookupRequest) (model.LookupResult, error) {
	resolvedSources, err := s.registry.SourcesFor(request)
	if err != nil {
		return model.LookupResult{}, err
	}

	result := model.LookupResult{
		Request:     request,
		GeneratedAt: time.Now().UTC(),
	}

	collected := collectOutcomes(ctx, request, resolvedSources)
	sortOutcomesByPriority(collected)

	for _, outcome := range collected {
		appendOutcome(&result, outcome)
	}

	_ = s.cacheResult(ctx, cacheKey, request, result)

	return result, nil
}

// Lookup fans out to all matching sources in parallel and merges results.
func (s *LookupService) Lookup(ctx context.Context, request model.LookupRequest) (model.LookupResult, error) {
	cacheKey := cache.BuildKey(request)
	cached, ok, err := s.lookupCachedResult(ctx, cacheKey, request)
	if err != nil {
		return model.LookupResult{}, err
	}
	if ok {
		return cached, nil
	}
	if request.NoCache {
		return s.lookupFreshResult(ctx, cacheKey, request)
	}

	result, _, err := s.flights.Do(ctx, cacheKey, func(callCtx context.Context) (model.LookupResult, error) {
		return s.lookupFreshResult(callCtx, cacheKey, request)
	})
	if err != nil {
		return model.LookupResult{}, err
	}

	return result, nil
}
