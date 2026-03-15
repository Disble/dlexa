package render

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/renderutil"
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

// RenderResult serializes the result as JSON, projecting article content when needed.
func (r *JSONRenderer) RenderResult(ctx context.Context, result model.LookupResult) ([]byte, error) {
	_ = ctx
	for idx := range result.Entries {
		if result.Entries[idx].Article != nil && result.Entries[idx].Content == "" {
			result.Entries[idx].Content = articleContentProjection(result.Entries[idx].Article)
		}
	}
	return marshalNoEscape(result)
}

func articleContentProjection(article *model.Article) string {
	if article == nil {
		return ""
	}
	var parts []string
	for _, section := range article.Sections {
		text := strings.TrimSpace(renderJSONSectionProjection(section))
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

func renderJSONSectionProjection(section model.Section) string {
	var parts []string
	heading := strings.TrimSpace(section.Label)
	if title := strings.TrimSpace(section.Title); title != "" {
		if heading != "" {
			heading += " "
		}
		heading += title
	}
	if heading != "" {
		parts = append(parts, heading)
	}
	for _, block := range sectionBlocks(section) {
		switch block.Kind {
		case model.ArticleBlockKindParagraph:
			if block.Paragraph != nil && strings.TrimSpace(block.Paragraph.Markdown) != "" {
				parts = append(parts, strings.TrimSpace(block.Paragraph.Markdown))
			}
		case model.ArticleBlockKindTable:
			if block.Table != nil {
				parts = append(parts, renderutil.RenderTableMarkdown(*block.Table, ""))
			}
		}
	}
	for _, child := range section.Children {
		if childText := strings.TrimSpace(renderJSONSectionProjection(child)); childText != "" {
			parts = append(parts, childText)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}
