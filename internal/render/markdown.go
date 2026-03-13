package render

import (
	"context"
	"fmt"
	"strings"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type MarkdownRenderer struct{}

func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

func (r *MarkdownRenderer) Format() string {
	return "markdown"
}

func (r *MarkdownRenderer) Render(ctx context.Context, result model.LookupResult) ([]byte, error) {
	return r.RenderResult(ctx, result)
}

func (r *MarkdownRenderer) RenderResult(ctx context.Context, result model.LookupResult) ([]byte, error) {
	_ = ctx
	var builder strings.Builder

	fmt.Fprintf(&builder, "# %s\n\n", result.Request.Query)
	fmt.Fprintf(&builder, "- format: `%s`\n", result.Request.Format)
	fmt.Fprintf(&builder, "- cache_hit: `%t`\n", result.CacheHit)
	fmt.Fprintf(&builder, "- sources: `%d`\n\n", len(result.Sources))

	for _, entry := range result.Entries {
		fmt.Fprintf(&builder, "## %s\n\n", entry.Headword)
		if entry.Summary != "" {
			fmt.Fprintf(&builder, "%s\n\n", entry.Summary)
		}
		fmt.Fprintf(&builder, "%s\n\n", entry.Content)
	}

	if len(result.Warnings) > 0 {
		builder.WriteString("## Warnings\n\n")
		for _, warning := range result.Warnings {
			fmt.Fprintf(&builder, "- [%s] %s (%s)\n", warning.Code, warning.Message, warning.Source)
		}
		builder.WriteString("\n")
	}

	if len(result.Problems) > 0 {
		builder.WriteString("## Problems\n\n")
		for _, problem := range result.Problems {
			fmt.Fprintf(&builder, "- [%s] %s (%s/%s)\n", problem.Code, problem.Message, problem.Source, problem.Severity)
		}
	}

	return []byte(builder.String()), nil
}
