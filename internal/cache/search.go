package cache

import (
	"context"
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
	return strings.Join([]string{"search", NormalizeSearchQuery(request.Query)}, "|")
}

// NormalizeSearchQuery compacts whitespace for search-query cache addressing.
func NormalizeSearchQuery(raw string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
}
