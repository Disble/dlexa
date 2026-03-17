package render

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

// SearchMarkdownRenderer renders search results as human-readable markdown/text.
type SearchMarkdownRenderer struct{}

// NewSearchMarkdownRenderer creates a SearchMarkdownRenderer.
func NewSearchMarkdownRenderer() *SearchMarkdownRenderer {
	return &SearchMarkdownRenderer{}
}

// Format returns "markdown".
func (r *SearchMarkdownRenderer) Format() string {
	return "markdown"
}

// Render formats the search result as human-readable markdown/text.
func (r *SearchMarkdownRenderer) Render(ctx context.Context, result model.SearchResult) ([]byte, error) {
	_ = ctx
	var builder strings.Builder
	query := strings.TrimSpace(result.Request.Query)
	if len(result.Candidates) == 0 {
		fmt.Fprintf(&builder, "No DPD entry candidates found for %q.", query)
		return []byte(builder.String()), nil
	}

	fmt.Fprintf(&builder, "Candidate DPD entries for %q:\n", query)
	for _, candidate := range result.Candidates {
		fmt.Fprintf(&builder, "- %s -> %s\n", candidate.DisplayText, candidate.ArticleKey)
	}
	return []byte(strings.TrimRight(builder.String(), "\n")), nil
}
