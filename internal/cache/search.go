package cache

import (
	"context"
	"sort"
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

// SearchStore abstracts cache read/write operations for normalized search results.
//
// Get returns either a hit, a miss, or a degraded miss when the backing cache
// cannot provide a usable entry. Runtime callers must degrade to a fresh search
// instead of surfacing cache failures. Set is best-effort and must not turn a
// successful fresh search into a user-visible failure.
type SearchStore interface {
	Get(ctx context.Context, key string) (model.SearchResult, bool, error)
	Set(ctx context.Context, key string, result model.SearchResult) error
}

// BuildSearchKey produces a deterministic format-neutral cache key from a SearchRequest.
func BuildSearchKey(request model.SearchRequest) string {
	providerNames := NormalizeSearchProviders(request.Sources)
	providers := strings.Join(providerNames, ",")
	if providers == "" {
		providers = "default"
	}

	return strings.Join([]string{"search/v2", NormalizeSearchQuery(request.Query), "providers=" + providers}, "|")
}

// BuildLegacySearchKey preserves the pre-v2 single-provider cache layout.
func BuildLegacySearchKey(request model.SearchRequest) string {
	return strings.Join([]string{"search", NormalizeSearchQuery(request.Query)}, "|")
}

// LegacySearchKey returns the legacy key only when the request targets exactly the default provider.
func LegacySearchKey(request model.SearchRequest, providerName, defaultProvider string) (string, bool) {
	providerName = strings.TrimSpace(providerName)
	defaultProvider = strings.TrimSpace(defaultProvider)
	if providerName == "" || defaultProvider == "" || providerName != defaultProvider {
		return "", false
	}

	providers := NormalizeSearchProviders(request.Sources)
	if len(providers) > 1 {
		return "", false
	}
	if len(providers) == 1 && providers[0] != providerName {
		return "", false
	}

	return BuildLegacySearchKey(request), true
}

// NormalizeSearchQuery compacts whitespace for search-query cache addressing.
func NormalizeSearchQuery(raw string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
}

// NormalizeSearchProviders compacts, deduplicates, and sorts provider names for cache addressing.
func NormalizeSearchProviders(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(raw))
	providers := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.Join(strings.Fields(strings.TrimSpace(item)), " ")
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		providers = append(providers, trimmed)
	}

	sort.Strings(providers)
	return providers
}
