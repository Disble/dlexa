package render

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
)

// SearchJSONRenderer renders search results as structured JSON.
type SearchJSONRenderer struct{}

// NewSearchJSONRenderer creates a SearchJSONRenderer.
func NewSearchJSONRenderer() *SearchJSONRenderer {
	return &SearchJSONRenderer{}
}

// Format returns "json".
func (r *SearchJSONRenderer) Format() string {
	return "json"
}

// Render serializes the search result as JSON.
func (r *SearchJSONRenderer) Render(ctx context.Context, result model.SearchResult) ([]byte, error) {
	_ = ctx
	return marshalNoEscape(result)
}
