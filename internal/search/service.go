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

const allSearchProvidersFailedMessage = "all search providers failed"

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
	registry        ProviderSelector
	cache           cache.SearchStore
	maxConcurrent   int
	defaultProvider string
	now             func() time.Time
	flights         inflight.Group[model.SearchResult]
}

// NewService creates a search service backed by the given provider registry and cache.
func NewService(registry ProviderSelector, store cache.SearchStore, maxConcurrent int, defaultProvider string) *Service {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	return &Service{registry: registry, cache: store, maxConcurrent: maxConcurrent, defaultProvider: defaultProvider, now: func() time.Time { return time.Now().UTC() }}
}

// RegistryForTesting exposes the wired registry for app wiring tests.
func (s *Service) RegistryForTesting() ProviderSelector { return s.registry }

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
		return allSearchProvidersFailedMessage
	}
	return allSearchProvidersFailedMessage + ": " + strings.Join(parts, "; ")
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

type aggregateState struct {
	result     model.SearchResult
	priorities map[string]int
	successes  int
}

func newAggregateState(baseRequest model.SearchRequest, providers []Provider, now time.Time, cached bool) aggregateState {
	state := aggregateState{
		result:     model.SearchResult{Request: aggregateRequest(baseRequest, providers), GeneratedAt: now, CacheHit: cached},
		priorities: make(map[string]int, len(providers)),
	}
	for _, provider := range providers {
		state.priorities[provider.Descriptor().Name] = provider.Descriptor().Priority
	}
	return state
}

func (s *aggregateState) addOutcome(outcome providerOutcome) {
	if outcome.err != nil {
		s.addFailure(outcome)
		return
	}
	if !outcome.cached {
		s.result.CacheHit = false
	}
	s.successes++
	s.appendWarnings(outcome)
	s.appendProblems(outcome)
	s.appendCandidates(outcome)
	s.updateGeneratedAt(outcome.result.GeneratedAt)
}

func (s *aggregateState) addFailure(outcome providerOutcome) {
	s.result.CacheHit = false
	s.result.Problems = append(s.result.Problems, problemFromError(outcome.provider, outcome.err))
}

func (s *aggregateState) appendWarnings(outcome providerOutcome) {
	for _, warning := range outcome.result.Warnings {
		if warning.Source == "" {
			warning.Source = outcome.provider.Descriptor().Name
		}
		s.result.Warnings = append(s.result.Warnings, warning)
	}
}

func (s *aggregateState) appendProblems(outcome providerOutcome) {
	for _, problem := range outcome.result.Problems {
		s.result.Problems = append(s.result.Problems, withProblemSource(problem, outcome.provider))
	}
}

func (s *aggregateState) appendCandidates(outcome providerOutcome) {
	for _, candidate := range outcome.result.Candidates {
		s.result.Candidates = append(s.result.Candidates, withCandidateSourceHint(candidate, outcome.provider))
	}
}

func (s *aggregateState) updateGeneratedAt(generatedAt time.Time) {
	if generatedAt.IsZero() {
		return
	}
	if s.result.GeneratedAt.IsZero() || generatedAt.Before(s.result.GeneratedAt) {
		s.result.GeneratedAt = generatedAt
	}
}

func (s *aggregateState) sortProblems() {
	sortProblemsByPriority(s.result.Problems, s.priorities)
}

func (s *aggregateState) finalize(now time.Time, outcomes []providerOutcome) (model.SearchResult, error) {
	if s.successes == 0 {
		return s.finalizeFailure(outcomes)
	}
	if s.result.GeneratedAt.IsZero() {
		s.result.GeneratedAt = now
	}
	return s.result, nil
}

func (s *aggregateState) finalizeFailure(outcomes []providerOutcome) (model.SearchResult, error) {
	if len(s.result.Problems) == 0 {
		return model.SearchResult{}, errors.New(allSearchProvidersFailedMessage)
	}
	if err := passthroughAggregateError(outcomes, s.result.Problems); err != nil {
		return model.SearchResult{}, err
	}
	if allProblemsRateLimited(s.result.Problems) {
		return model.SearchResult{}, newAggregateProblemError(
			model.ProblemCodeSearchAllProvidersRateLimited,
			aggregateProblemMessage(s.result.Problems),
			"all search providers rate limited",
		)
	}
	return model.SearchResult{}, newAggregateProblemError(
		aggregateProblemCode(s.result.Problems),
		aggregateProblemMessage(s.result.Problems),
		allSearchProvidersFailedMessage,
	)
}

func passthroughAggregateError(outcomes []providerOutcome, problems []model.Problem) error {
	if len(problems) > 1 {
		return nil
	}
	err := firstOutcomeError(outcomes)
	if err == nil {
		return nil
	}
	if len(outcomes) == 1 {
		return err
	}
	problem, ok := model.AsProblem(err)
	if !ok {
		return err
	}
	problem.Message = aggregateProblemMessage(problems)
	problem.Source = ""
	return model.NewProblemError(problem, errors.New(allSearchProvidersFailedMessage))
}

func newAggregateProblemError(code, message, cause string) error {
	return model.NewProblemError(
		model.Problem{Code: code, Message: message, Severity: model.ProblemSeverityError},
		errors.New(cause),
	)
}

func aggregateResults(baseRequest model.SearchRequest, providers []Provider, outcomes []providerOutcome, now time.Time) (model.SearchResult, error) {
	state := newAggregateState(baseRequest, providers, now, len(outcomes) > 0)
	for _, outcome := range outcomes {
		state.addOutcome(outcome)
	}
	state.sortProblems()
	return state.finalize(now, outcomes)
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
