package cache

import (
	"context"
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

// SearchStore abstracts cache read/write operations for normalized search results.
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
