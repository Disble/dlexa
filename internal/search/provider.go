package search

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// Provider executes one concrete search pipeline for a source.
type Provider interface {
	Descriptor() model.SourceDescriptor
	Search(ctx context.Context, request model.SearchRequest) (model.SearchResult, error)
}

// Registry selects providers for a given search request.
type Registry interface {
	ProvidersFor(request model.SearchRequest) ([]Provider, error)
}

// PipelineProvider adapts an existing fetch/parse/normalize chain into a search provider.
type PipelineProvider struct {
	descriptor model.SourceDescriptor
	fetcher    fetch.Fetcher
	parser     Parser
	normalizer Normalizer
	now        func() time.Time
}

// NewPipelineProvider creates a provider backed by the existing live-search pipeline.
func NewPipelineProvider(descriptor model.SourceDescriptor, fetcher fetch.Fetcher, parser Parser, normalizer Normalizer) *PipelineProvider {
	return &PipelineProvider{
		descriptor: descriptor,
		fetcher:    fetcher,
		parser:     parser,
		normalizer: normalizer,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

// Descriptor returns the provider identity.
func (p *PipelineProvider) Descriptor() model.SourceDescriptor { return p.descriptor }

// FetcherForTesting exposes the wired fetcher for wiring tests.
func (p *PipelineProvider) FetcherForTesting() fetch.Fetcher { return p.fetcher }

// ParserForTesting exposes the wired parser for wiring tests.
func (p *PipelineProvider) ParserForTesting() Parser { return p.parser }

// NormalizerForTesting exposes the wired normalizer for wiring tests.
func (p *PipelineProvider) NormalizerForTesting() Normalizer { return p.normalizer }

// Search executes the provider pipeline without cache concerns.
func (p *PipelineProvider) Search(ctx context.Context, request model.SearchRequest) (model.SearchResult, error) {
	if p == nil {
		return model.SearchResult{}, fmt.Errorf("search provider is not configured")
	}

	document, err := p.fetcher.Fetch(ctx, fetch.Request{Query: request.Query, Source: p.descriptor})
	if err != nil {
		return model.SearchResult{}, err
	}
	records, parseWarnings, err := p.parser.Parse(ctx, p.descriptor, document)
	if err != nil {
		return model.SearchResult{}, err
	}
	candidates, normalizeWarnings, err := p.normalizer.Normalize(ctx, p.descriptor, records)
	if err != nil {
		return model.SearchResult{}, err
	}

	return model.SearchResult{
		Request:     request,
		Candidates:  candidates,
		Warnings:    append(append([]model.Warning(nil), parseWarnings...), normalizeWarnings...),
		GeneratedAt: p.now(),
	}, nil
}

// StaticRegistry stores a fixed, priority-ordered provider set.
type StaticRegistry struct {
	providers       []Provider
	defaultProvider string
}

// NewStaticRegistry creates a registry ordered by provider priority.
func NewStaticRegistry(defaultProvider string, providers ...Provider) *StaticRegistry {
	ordered := append([]Provider(nil), providers...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Descriptor().Priority < ordered[j].Descriptor().Priority
	})

	return &StaticRegistry{providers: ordered, defaultProvider: strings.TrimSpace(defaultProvider)}
}

// Providers returns a snapshot of registered providers for wiring tests.
func (r *StaticRegistry) Providers() []Provider {
	if r == nil {
		return nil
	}
	return append([]Provider(nil), r.providers...)
}

// DefaultProviderForTesting exposes the configured default provider for wiring tests.
func (r *StaticRegistry) DefaultProviderForTesting() string {
	if r == nil {
		return ""
	}
	return r.defaultProvider
}

// ProvidersFor selects providers by request or falls back to the configured default provider.
func (r *StaticRegistry) ProvidersFor(request model.SearchRequest) ([]Provider, error) {
	if r == nil || len(r.providers) == 0 {
		return nil, fmt.Errorf("no search providers configured")
	}

	allowed := requestedProviders(request.Sources)
	if len(allowed) == 0 {
		defaultName := r.defaultProvider
		if defaultName == "" {
			defaultName = r.providers[0].Descriptor().Name
		}
		allowed = map[string]struct{}{defaultName: {}}
	}

	selected := make([]Provider, 0, len(allowed))
	for _, provider := range r.providers {
		if _, ok := allowed[provider.Descriptor().Name]; ok {
			selected = append(selected, provider)
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no search providers matched request: %v", request.Sources)
	}
	return selected, nil
}

func requestedProviders(raw []string) map[string]struct{} {
	if len(raw) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		allowed[trimmed] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil
	}
	return allowed
}

// Ensure ParsedSearchRecord stays referenced from this package for provider adapters.
var _ []parse.ParsedSearchRecord
