package render

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

// JSONRenderer renders lookup results as indented JSON.
type JSONRenderer struct{}

// NewJSONRenderer creates a JSONRenderer.
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

// Format returns "json".
func (r *JSONRenderer) Format() string {
	return "json"
}

// Render serializes the result as JSON.
func (r *JSONRenderer) Render(ctx context.Context, result model.LookupResult) ([]byte, error) {
	return r.RenderResult(ctx, result)
}

// RenderResult serializes the result as JSON, projecting article content from the
// structured article whenever available so JSON compatibility fields stay aligned
// with the current Markdown-facing semantics.
func (r *JSONRenderer) RenderResult(ctx context.Context, result model.LookupResult) ([]byte, error) {
	_ = ctx
	for idx := range result.Entries {
		if result.Entries[idx].Article != nil {
			result.Entries[idx].Content = articleContentProjection(result.Entries[idx].Article)
		}
	}
	if len(result.Misses) == 0 {
		result.Misses = nil
	}
	return marshalNoEscape(result)
}

func articleContentProjection(article *model.Article) string {
	if article == nil {
		return ""
	}
	renderer := NewMarkdownRenderer()
	var parts []string
	for _, section := range article.Sections {
		text := strings.TrimSpace(renderer.renderMarkdownSection(section, ""))
		if text == "" {
			continue
		}
		parts = append(parts, text)
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func marshalNoEscape(v any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}
