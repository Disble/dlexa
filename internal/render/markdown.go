package render

import (
	"context"
	"fmt"
	"strings"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type MarkdownRenderer struct {
	profile TerminalProfile
}

func NewMarkdownRenderer() *MarkdownRenderer {
	return NewMarkdownRendererWithProfile(DefaultTerminalProfile())
}

func NewMarkdownRendererWithProfile(profile TerminalProfile) *MarkdownRenderer {
	return &MarkdownRenderer{profile: normalizeTerminalProfile(profile)}
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

	if areAllArticleEntries(result.Entries) {
		return r.renderArticleGroupMarkdown(result), nil
	}

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

func areAllArticleEntries(entries []model.Entry) bool {
	if len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		if entry.Article == nil {
			return false
		}
	}
	return true
}

func (r *MarkdownRenderer) renderArticleGroupMarkdown(result model.LookupResult) []byte {
	var builder strings.Builder
	heading := strings.TrimSpace(result.Request.Query)
	if heading == "" && len(result.Entries) > 0 {
		heading = result.Entries[0].Headword
	}
	multiple := len(result.Entries) > 1
	if multiple {
		fmt.Fprintf(&builder, "# %s\n\n", heading)
	}

	for idx, entry := range result.Entries {
		if idx > 0 {
			builder.WriteString("\n")
		}
		builder.Write(r.renderEntryArticleMarkdown(entry, multiple))
		if idx < len(result.Entries)-1 {
			builder.WriteString("\n")
		}
	}

	return []byte(strings.TrimSpace(builder.String()) + "\n")
}

func (r *MarkdownRenderer) renderEntryArticleMarkdown(entry model.Entry, grouped bool) []byte {
	article := entry.Article
	var builder strings.Builder
	if grouped {
		heading := strings.TrimSpace(entry.Headword)
		if heading == "" {
			heading = strings.TrimSpace(article.Lemma)
		}
		if heading == "" {
			heading = entry.ID
		}
		fmt.Fprintf(&builder, "## %s\n\n", heading)
	} else {
		fmt.Fprintf(&builder, "# %s\n\n", entry.Headword)
	}
	if article.Dictionary != "" {
		fmt.Fprintf(&builder, "%s\n\n", article.Dictionary)
	}
	if article.Edition != "" {
		fmt.Fprintf(&builder, "%s\n\n", article.Edition)
	}
	for idx, section := range article.Sections {
		sectionText := strings.TrimSpace(r.renderMarkdownSection(section, ""))
		if sectionText == "" {
			continue
		}
		if idx > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(sectionText)
	}
	if citation := strings.TrimSpace(renderCitationMarkdown(article)); citation != "" {
		builder.WriteString("\n\n")
		builder.WriteString(citation)
	}
	return []byte(strings.TrimSpace(builder.String()))
}

func (r *MarkdownRenderer) renderMarkdownSection(section model.Section, indent string) string {
	var builder strings.Builder
	heading := strings.TrimSpace(section.Label)
	if title := strings.TrimSpace(section.Title); title != "" {
		if heading != "" {
			heading += " "
		}
		heading += title
	}

	blocks := sectionBlocks(section)
	if heading != "" {
		builder.WriteString(indent)
		builder.WriteString(heading)
	}

	firstParagraphInline := false
	for _, block := range blocks {
		switch block.Kind {
		case model.ArticleBlockKindParagraph:
			if block.Paragraph == nil {
				continue
			}
			text := strings.TrimSpace(r.renderMarkdownParagraph(*block.Paragraph))
			if text == "" {
				continue
			}
			if heading != "" && !firstParagraphInline {
				builder.WriteString(" ")
				builder.WriteString(text)
				firstParagraphInline = true
				continue
			}
			if builder.Len() > 0 {
				builder.WriteString("\n\n")
			}
			builder.WriteString(indentLines(text, indent))
		case model.ArticleBlockKindTable:
			if block.Table == nil {
				continue
			}
			tableText := renderTableMarkdown(*block.Table, indent)
			if tableText == "" {
				continue
			}
			if builder.Len() > 0 {
				builder.WriteString("\n\n")
			}
			builder.WriteString(tableText)
		}
	}

	for _, child := range section.Children {
		childText := strings.TrimRight(r.renderMarkdownSection(child, indent+"  "), "\n")
		if childText == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(childText)
	}
	return strings.TrimRight(builder.String(), "\n")
}

func sectionBlocks(section model.Section) []model.Block {
	if len(section.Blocks) > 0 {
		return section.Blocks
	}
	blocks := make([]model.Block, 0, len(section.Paragraphs))
	for _, paragraph := range section.Paragraphs {
		paragraphCopy := paragraph
		blocks = append(blocks, model.Block{Kind: model.ArticleBlockKindParagraph, Paragraph: &paragraphCopy})
	}
	return blocks
}

func (r *MarkdownRenderer) renderMarkdownParagraph(paragraph model.Paragraph) string {
	if len(paragraph.Inlines) > 0 {
		return renderMarkdownInlines(paragraph.Inlines)
	}
	return normalizeLegacyMarkdownProjection(paragraph.Markdown)
}

func renderMarkdownInlines(inlines []model.Inline) string {
	var builder strings.Builder
	for _, inline := range inlines {
		piece := renderMarkdownInline(inline)
		if piece == "" {
			continue
		}
		if builder.Len() > 0 && needsInlineSpace(builder.String(), piece) {
			builder.WriteString(" ")
		}
		builder.WriteString(piece)
	}
	return strings.TrimSpace(builder.String())
}

func renderMarkdownInline(inline model.Inline) string {
	text := strings.TrimSpace(inline.Text)
	if len(inline.Children) > 0 {
		text = renderMarkdownInlines(inline.Children)
	}
	if text == "" {
		return ""
	}

	switch inline.Kind {
	case model.InlineKindExample,
		model.InlineKindMention,
		model.InlineKindEmphasis,
		model.InlineKindWorkTitle,
		model.InlineKindCorrection:
		return "*" + text + "*"
	case model.InlineKindReference:
		return "→ [" + text + "](" + inline.Target + ")"
	default:
		return text
	}
}

func needsInlineSpace(current, next string) bool {
	if current == "" || next == "" {
		return false
	}
	last := rune(current[len(current)-1])
	first := rune(next[0])
	if strings.ContainsRune(" [{«", first) || strings.ContainsRune(")]}.;,!?»:", first) {
		return false
	}
	if strings.ContainsRune(" ([{«", last) {
		return false
	}
	return true
}

func normalizeLegacyMarkdownProjection(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "‹", "*")
	text = strings.ReplaceAll(text, "›", "*")
	return strings.TrimSpace(text)
}

func cleanTerminalProjection(raw string) string {
	return normalizeLegacyMarkdownProjection(raw)
}

func indentLines(text string, indent string) string {
	if indent == "" || strings.TrimSpace(text) == "" {
		return text
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

func renderTableMarkdown(table model.Table, indent string) string {
	rows := make([][]string, 0, len(table.Headers)+len(table.Rows))
	for _, row := range table.Headers {
		rows = append(rows, tableRowTexts(row))
	}
	for _, row := range table.Rows {
		rows = append(rows, tableRowTexts(row))
	}
	if len(rows) == 0 {
		return ""
	}
	widths := tableColumnWidths(rows)
	lines := make([]string, 0, len(rows)+1)
	for idx, row := range rows {
		lines = append(lines, indent+formatTableRow(row, widths))
		if idx == len(table.Headers)-1 && len(table.Headers) > 0 {
			lines = append(lines, indent+formatTableDivider(widths))
		}
	}
	return strings.Join(lines, "\n")
}

func tableRowTexts(row model.TableRow) []string {
	result := make([]string, 0, len(row.Cells))
	for _, cell := range row.Cells {
		result = append(result, cell.Text)
	}
	return result
}

func tableColumnWidths(rows [][]string) []int {
	widths := make([]int, 0)
	for _, row := range rows {
		for idx, cell := range row {
			if idx >= len(widths) {
				widths = append(widths, len(cell))
				continue
			}
			if len(cell) > widths[idx] {
				widths[idx] = len(cell)
			}
		}
	}
	for idx := range widths {
		if widths[idx] == 0 {
			widths[idx] = 1
		}
	}
	return widths
}

func formatTableRow(row []string, widths []int) string {
	parts := make([]string, 0, len(widths))
	for idx, width := range widths {
		cell := ""
		if idx < len(row) {
			cell = row[idx]
		}
		parts = append(parts, fmt.Sprintf(" %-*s ", width, cell))
	}
	return "|" + strings.Join(parts, "|") + "|"
}

func formatTableDivider(widths []int) string {
	parts := make([]string, 0, len(widths))
	for _, width := range widths {
		parts = append(parts, strings.Repeat("-", width+2))
	}
	return "|" + strings.Join(parts, "+") + "|"
}

func renderCitationMarkdown(article *model.Article) string {
	if article == nil {
		return ""
	}

	citation := article.Citation
	lines := make([]string, 0, 5)
	if value := strings.TrimSpace(citation.SourceLabel); value != "" {
		lines = append(lines, "Source: "+value)
	}
	if len(lines) > 0 {
		if value := strings.TrimSpace(article.Dictionary); value != "" {
			lines = append(lines, "Dictionary: "+value)
		}
		if value := strings.TrimSpace(citation.Edition); value != "" {
			lines = append(lines, "Edition: "+value)
		}
		if value := strings.TrimSpace(citation.CanonicalURL); value != "" {
			lines = append(lines, "URL: "+value)
		}
		if value := strings.TrimSpace(citation.ConsultedAt); value != "" {
			lines = append(lines, "Consulted: "+value)
		}
		return strings.Join(lines, "\n")
	}

	if value := strings.TrimSpace(article.Dictionary); value != "" {
		lines = append(lines, "Dictionary: "+value)
	}
	if len(lines) > 0 && strings.TrimSpace(citation.Text) == "" {
		return strings.Join(lines, "\n")
	}

	return strings.TrimSpace(citation.Text)
}
