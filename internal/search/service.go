package search

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Disble/dlexa/internal/cache"
	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/inflight"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// Searcher defines the contract for performing semantic search queries.
type Searcher interface {
	Search(ctx context.Context, request model.SearchRequest) (model.SearchResult, error)
}

// Parser decodes fetched search payloads into parsed records.
type Parser interface {
	Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]parse.ParsedSearchRecord, []model.Warning, error)
}

// Normalizer converts parsed search records into normalized search candidates.
type Normalizer interface {
	Normalize(ctx context.Context, descriptor model.SourceDescriptor, records []parse.ParsedSearchRecord) ([]model.SearchCandidate, []model.Warning, error)
}

// Service orchestrates cache-aside semantic search execution.
type Service struct {
	registry        Registry
	cache           cache.SearchStore
	maxConcurrent   int
	defaultProvider string
	now             func() time.Time
	flights         inflight.Group[model.SearchResult]
}

// NewService creates a search service backed by the given provider registry and cache.
func NewService(registry Registry, store cache.SearchStore, maxConcurrent int, defaultProvider string) *Service {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	return &Service{registry: registry, cache: store, maxConcurrent: maxConcurrent, defaultProvider: defaultProvider, now: func() time.Time { return time.Now().UTC() }}
}

// RegistryForTesting exposes the wired registry for app wiring tests.
func (s *Service) RegistryForTesting() Registry { return s.registry }

// MaxConcurrentForTesting exposes the configured concurrency bound for wiring tests.
func (s *Service) MaxConcurrentForTesting() int { return s.maxConcurrent }

func (s *Service) cacheResult(ctx context.Context, cacheKey string, request model.SearchRequest, result model.SearchResult) error {
	if s.cache == nil || request.NoCache {
		return nil
	}

	cacheCopy := result
	cacheCopy.Request = model.SearchRequest{Query: cache.NormalizeSearchQuery(request.Query), Sources: append([]string(nil), request.Sources...)}

	return s.cache.Set(ctx, cacheKey, cacheCopy)
}

func (s *Service) lookupCachedResult(ctx context.Context, provider Provider, request model.SearchRequest) (model.SearchResult, bool) {
	if s.cache == nil || request.NoCache {
		return model.SearchResult{}, false
	}
	cacheKey := cache.BuildSearchKey(request)

	cached, ok, err := s.cache.Get(ctx, cacheKey)
	if (err != nil || !ok) && s.cache != nil {
		if legacyKey, legacyOK := cache.LegacySearchKey(request, provider.Descriptor().Name, s.defaultProvider); legacyOK {
			legacyCached, legacyHit, legacyErr := s.cache.Get(ctx, legacyKey)
			if legacyErr == nil && legacyHit {
				cached, ok, err = legacyCached, true, nil
				_ = s.cache.Set(ctx, cacheKey, legacyCached)
			}
		}
	}
	if err != nil || !ok {
		return model.SearchResult{}, false
	}

	cached.Request = request
	cached.CacheHit = true
	return cached, true
}

func providerRequest(base model.SearchRequest, provider Provider) model.SearchRequest {
	request := base
	request.Sources = []string{provider.Descriptor().Name}
	return request
}

func withProblemSource(problem model.Problem, provider Provider) model.Problem {
	if problem.Source == "" {
		problem.Source = provider.Descriptor().Name
	}
	return problem
}

func withCandidateSourceHint(candidate model.SearchCandidate, provider Provider) model.SearchCandidate {
	if strings.TrimSpace(candidate.SourceHint) != "" {
		return candidate
	}
	descriptor := provider.Descriptor()
	if display := strings.TrimSpace(descriptor.DisplayName); display != "" {
		candidate.SourceHint = display
		return candidate
	}
	candidate.SourceHint = strings.TrimSpace(descriptor.Name)
	return candidate
}

func problemFromError(provider Provider, err error) model.Problem {
	problem, ok := model.AsProblem(err)
	if !ok {
		return model.Problem{Code: model.ProblemCodeSourceLookupFailed, Message: err.Error(), Source: provider.Descriptor().Name, Severity: model.ProblemSeverityError}
	}
	return withProblemSource(problem, provider)
}

func sortProblemsByPriority(problems []model.Problem, priorities map[string]int) {
	sort.SliceStable(problems, func(i, j int) bool {
		return priorities[problems[i].Source] < priorities[problems[j].Source]
	})
}

type providerOutcome struct {
	provider Provider
	result   model.SearchResult
	err      error
	cached   bool
}

func firstOutcomeError(outcomes []providerOutcome) error {
	for _, outcome := range outcomes {
		if outcome.err != nil {
			return outcome.err
		}
	}
	return nil
}

func aggregateProblemCode(problems []model.Problem) string {
	if allProblemsRateLimited(problems) {
		return model.ProblemCodeSearchAllProvidersRateLimited
	}
	for _, problem := range problems {
		switch problem.Code {
		case model.ProblemCodeDPDSearchParseFailed, model.ProblemCodeDPDSearchNormalizeFailed, model.ProblemCodeDPDExtractFailed, model.ProblemCodeDPDTransformFailed:
			return problem.Code
		}
	}
	for _, problem := range problems {
		if strings.TrimSpace(problem.Code) != "" {
			return problem.Code
		}
	}
	return model.ProblemCodeSourceLookupFailed
}

func allProblemsRateLimited(problems []model.Problem) bool {
	if len(problems) == 0 {
		return false
	}
	for _, problem := range problems {
		if problem.Code != model.ProblemCodeDPDSearchRateLimited {
			return false
		}
	}
	return true
}

func aggregateProblemMessage(problems []model.Problem) string {
	parts := make([]string, 0, len(problems))
	for _, problem := range problems {
		source := strings.TrimSpace(problem.Source)
		message := strings.TrimSpace(problem.Message)
		if message == "" {
			message = strings.TrimSpace(problem.Code)
		}
		if message == "" {
			message = "unknown provider failure"
		}
		if source == "" {
			parts = append(parts, message)
			continue
		}
		parts = append(parts, "["+source+"] "+message)
	}
	if len(parts) == 0 {
		return "all search providers failed"
	}
	return "all search providers failed: " + strings.Join(parts, "; ")
}

func (s *Service) searchProvider(ctx context.Context, provider Provider, baseRequest model.SearchRequest) providerOutcome {
	request := providerRequest(baseRequest, provider)
	if cached, ok := s.lookupCachedResult(ctx, provider, request); ok {
		return providerOutcome{provider: provider, result: cached, cached: true}
	}

	cacheKey := cache.BuildSearchKey(request)
	if request.NoCache {
		result, err := provider.Search(ctx, request)
		if err != nil {
			return providerOutcome{provider: provider, err: err}
		}
		result.Request = request
		result.CacheHit = false
		return providerOutcome{provider: provider, result: result}
	}

	result, _, err := s.flights.Do(ctx, cacheKey, func(callCtx context.Context) (model.SearchResult, error) {
		fresh, searchErr := provider.Search(callCtx, request)
		if searchErr != nil {
			return model.SearchResult{}, searchErr
		}
		fresh.Request = request
		fresh.CacheHit = false
		_ = s.cacheResult(callCtx, cacheKey, request, fresh)
		return fresh, nil
	})
	if err != nil {
		return providerOutcome{provider: provider, err: err}
	}
	result.Request = request
	result.CacheHit = false
	return providerOutcome{provider: provider, result: result}
}

func (s *Service) collectOutcomes(ctx context.Context, providers []Provider, request model.SearchRequest) []providerOutcome {
	outcomes := make(chan providerOutcome, len(providers))
	sem := make(chan struct{}, s.maxConcurrent)
	var wg sync.WaitGroup

	for _, provider := range providers {
		provider := provider
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				outcomes <- providerOutcome{provider: provider, err: ctx.Err()}
				return
			}
			defer func() { <-sem }()
			outcomes <- s.searchProvider(ctx, provider, request)
		}()
	}

	go func() {
		wg.Wait()
		close(outcomes)
	}()

	collected := make([]providerOutcome, 0, len(providers))
	for outcome := range outcomes {
		collected = append(collected, outcome)
	}
	sort.SliceStable(collected, func(i, j int) bool {
		return collected[i].provider.Descriptor().Priority < collected[j].provider.Descriptor().Priority
	})
	return collected
}

func aggregateRequest(baseRequest model.SearchRequest, providers []Provider) model.SearchRequest {
	request := baseRequest
	request.Sources = make([]string, 0, len(providers))
	for _, provider := range providers {
		request.Sources = append(request.Sources, provider.Descriptor().Name)
	}
	return request
}

func aggregateResults(baseRequest model.SearchRequest, providers []Provider, outcomes []providerOutcome, now time.Time) (model.SearchResult, error) {
	result := model.SearchResult{Request: aggregateRequest(baseRequest, providers), GeneratedAt: now, CacheHit: len(outcomes) > 0}
	priorities := make(map[string]int, len(providers))
	successes := 0

	for _, provider := range providers {
		priorities[provider.Descriptor().Name] = provider.Descriptor().Priority
	}

	for _, outcome := range outcomes {
		if outcome.err != nil {
			result.CacheHit = false
			result.Problems = append(result.Problems, problemFromError(outcome.provider, outcome.err))
			continue
		}
		successes++
		if !outcome.cached {
			result.CacheHit = false
		}
		for _, warning := range outcome.result.Warnings {
			if warning.Source == "" {
				warning.Source = outcome.provider.Descriptor().Name
			}
			result.Warnings = append(result.Warnings, warning)
		}
		for _, problem := range outcome.result.Problems {
			result.Problems = append(result.Problems, withProblemSource(problem, outcome.provider))
		}
		for _, candidate := range outcome.result.Candidates {
			result.Candidates = append(result.Candidates, withCandidateSourceHint(candidate, outcome.provider))
		}
		if result.GeneratedAt.IsZero() || (!outcome.result.GeneratedAt.IsZero() && outcome.result.GeneratedAt.Before(result.GeneratedAt)) {
			result.GeneratedAt = outcome.result.GeneratedAt
		}
	}

	sortProblemsByPriority(result.Problems, priorities)
	if successes == 0 {
		if len(result.Problems) == 0 {
			return model.SearchResult{}, errors.New("all search providers failed")
		}
		if len(outcomes) == 1 {
			if err := firstOutcomeError(outcomes); err != nil {
				return model.SearchResult{}, err
			}
		}
		if len(result.Problems) == 1 {
			if err := firstOutcomeError(outcomes); err != nil {
				problem, ok := model.AsProblem(err)
				if ok {
					problem.Message = aggregateProblemMessage(result.Problems)
					problem.Source = ""
					return model.SearchResult{}, model.NewProblemError(problem, errors.New("all search providers failed"))
				}
				return model.SearchResult{}, err
			}
		}
		if allProblemsRateLimited(result.Problems) {
			return model.SearchResult{}, model.NewProblemError(model.Problem{Code: model.ProblemCodeSearchAllProvidersRateLimited, Message: aggregateProblemMessage(result.Problems), Severity: model.ProblemSeverityError}, errors.New("all search providers rate limited"))
		}
		return model.SearchResult{}, model.NewProblemError(model.Problem{Code: aggregateProblemCode(result.Problems), Message: aggregateProblemMessage(result.Problems), Severity: model.ProblemSeverityError}, errors.New("all search providers failed"))
	}
	if result.GeneratedAt.IsZero() {
		result.GeneratedAt = now
	}
	return result, nil
}

// Search runs a semantic search using a format-neutral cached normalized result when available.
func (s *Service) Search(ctx context.Context, request model.SearchRequest) (model.SearchResult, error) {
	providers, err := s.registry.ProvidersFor(request)
	if err != nil {
		return model.SearchResult{}, err
	}
	outcomes := s.collectOutcomes(ctx, providers, request)
	result, err := aggregateResults(request, providers, outcomes, s.now())
	if err != nil {
		return model.SearchResult{}, err
	}
	return result, nil
}
