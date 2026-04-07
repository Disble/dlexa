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
		fmt.Fprintf(&builder, "## Resultado semántico para %q\n\n", query)
		fmt.Fprintf(&builder, "No se encontraron rutas normativas accionables para %q.\n\n", query)
		builder.WriteString("- siguiente_paso: `dlexa search <consulta más específica>`")
		return []byte(builder.String()), nil
	}

	fmt.Fprintf(&builder, "## Resultado semántico para %q\n\n", query)
	fmt.Fprintf(&builder, "- total_candidatos: %d\n", len(result.Candidates))
	fmt.Fprintf(&builder, "- siguiente_paso: `%s`\n", searchNextCommand(result.Candidates[0], query))

	for idx, candidate := range result.Candidates {
		fmt.Fprintf(&builder, "\n### %d. %s\n", idx+1, searchTitle(candidate))
		fmt.Fprintf(&builder, "- snippet: %s\n", searchSnippet(candidate))
		if classification := strings.TrimSpace(candidate.Classification); classification != "" {
			fmt.Fprintf(&builder, "- clasificación: %s\n", classification)
		}
		if sourceHint := strings.TrimSpace(candidate.SourceHint); sourceHint != "" {
			fmt.Fprintf(&builder, "- fuente: %s\n", sourceHint)
		}
		fmt.Fprintf(&builder, "- next_command: `%s`\n", searchNextCommand(candidate, query))
	}
	return []byte(strings.TrimRight(builder.String(), "\n")), nil
}

func searchTitle(candidate model.SearchCandidate) string {
	for _, value := range []string{candidate.Title, candidate.DisplayText, candidate.ArticleKey, candidate.ID} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return "resultado"
}

func searchSnippet(candidate model.SearchCandidate) string {
	for _, value := range []string{candidate.Snippet, candidate.DisplayText, candidate.ArticleKey} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return "Sin snippet disponible."
}

func searchNextCommand(candidate model.SearchCandidate, query string) string {
	if trimmed := strings.TrimSpace(candidate.NextCommand); trimmed != "" {
		return trimmed
	}
	if key := strings.TrimSpace(candidate.ArticleKey); key != "" {
		return fmt.Sprintf("dlexa dpd %s", key)
	}
	if trimmedQuery := strings.TrimSpace(query); trimmedQuery != "" {
		return fmt.Sprintf("dlexa search %s", trimmedQuery)
	}
	return "dlexa search <consulta>"
}
