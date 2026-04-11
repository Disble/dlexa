package render

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/renderutil"
)

const (
	mdFmtH1    = "# %s\n\n"
	mdFmtBlock = "%s\n\n"
)

// MarkdownRenderer renders lookup results as Markdown text.
type MarkdownRenderer struct {
	profile TerminalProfile
}

// NewMarkdownRenderer creates a MarkdownRenderer with the default terminal profile.
func NewMarkdownRenderer() *MarkdownRenderer {
	return NewMarkdownRendererWithProfile(DefaultTerminalProfile())
}

// NewMarkdownRendererWithProfile creates a MarkdownRenderer using the given terminal profile.
func NewMarkdownRendererWithProfile(profile TerminalProfile) *MarkdownRenderer {
	return &MarkdownRenderer{profile: normalizeTerminalProfile(profile)}
}

// Format returns "markdown".
func (r *MarkdownRenderer) Format() string {
	return "markdown"
}

// Render formats the result as Markdown.
func (r *MarkdownRenderer) Render(ctx context.Context, result model.LookupResult) ([]byte, error) {
	return r.RenderResult(ctx, result)
}

// RenderResult formats the result as Markdown, using article rendering when all entries have articles.
func (r *MarkdownRenderer) RenderResult(ctx context.Context, result model.LookupResult) ([]byte, error) {
	_ = ctx
	var builder strings.Builder

	if areAllArticleEntries(result.Entries) {
		return r.renderArticleGroupMarkdown(result), nil
	}
	if len(result.Entries) == 0 && len(result.Misses) > 0 {
		return []byte(renderLookupMissMarkdown(result)), nil
	}

	fmt.Fprintf(&builder, mdFmtH1, result.Request.Query)
	fmt.Fprintf(&builder, "- format: `%s`\n", result.Request.Format)
	fmt.Fprintf(&builder, "- cache_hit: `%t`\n", result.CacheHit)
	fmt.Fprintf(&builder, "- sources: `%d`\n\n", len(result.Sources))

	for _, entry := range result.Entries {
		fmt.Fprintf(&builder, "## %s\n\n", entry.Headword)
		if entry.Summary != "" {
			fmt.Fprintf(&builder, mdFmtBlock, entry.Summary)
		}
		fmt.Fprintf(&builder, mdFmtBlock, entry.Content)
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

func renderLookupMissMarkdown(result model.LookupResult) string {
	var builder strings.Builder
	query := strings.TrimSpace(result.Request.Query)
	if query == "" && len(result.Misses) > 0 {
		query = result.Misses[0].Query
	}
	if query == "" {
		query = "lookup"
	}
	fmt.Fprintf(&builder, mdFmtH1, query)
	for idx, miss := range result.Misses {
		if idx > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(renderSingleLookupMissMarkdown(miss))
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String()) + "\n"
}

func renderSingleLookupMissMarkdown(miss model.LookupMiss) string {
	if miss.Suggestion != nil {
		if suggestionURL := strings.TrimSpace(miss.Suggestion.URL); suggestionURL != "" {
			return fmt.Sprintf("Quizá quiso decir **%s**.\n\n%s", miss.Suggestion.DisplayText, suggestionURL)
		}
		return fmt.Sprintf("Quizá quiso decir **%s**.", miss.Suggestion.DisplayText)
	}
	if miss.NextAction != nil && strings.TrimSpace(miss.NextAction.Command) != "" {
		return fmt.Sprintf("Try `%s`.", miss.NextAction.Command)
	}
	if strings.TrimSpace(miss.NoticeText) != "" {
		return miss.NoticeText
	}
	return "No exact DPD entry found."
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
	heading := articleGroupHeading(result)
	multiple := len(result.Entries) > 1
	if multiple {
		fmt.Fprintf(&builder, mdFmtH1, heading)
	}

	appendRedirectWarnings(&builder, result.Warnings, result.Request.Query)
	redirectedFromURL := redirectedSourceURL(result.Warnings)
	for idx, entry := range result.Entries {
		appendRenderedArticleEntry(&builder, r.renderEntryArticleMarkdown(entry, multiple, redirectedFromURL), idx, len(result.Entries))
	}

	return []byte(strings.TrimSpace(builder.String()) + "\n")
}

func articleGroupHeading(result model.LookupResult) string {
	heading := strings.TrimSpace(result.Request.Query)
	if heading == "" && len(result.Entries) > 0 {
		return result.Entries[0].Headword
	}
	return heading
}

func appendRedirectWarnings(builder *strings.Builder, warnings []model.Warning, query string) {
	for _, warning := range warnings {
		if warning.Code != model.WarningCodeDPDRedirected {
			continue
		}
		fmt.Fprintf(builder, "> ⚠ El DPD redirige %s\n>\n> El contenido mostrado corresponde a la entrada de destino, no a \"%s\".\n\n", warning.Message, query)
	}
}

func redirectedSourceURL(warnings []model.Warning) string {
	for _, warning := range warnings {
		if warning.Code != model.WarningCodeDPDRedirected {
			continue
		}
		parts := strings.SplitN(warning.Message, " → ", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0])
		}
		break
	}
	return ""
}

func appendRenderedArticleEntry(builder *strings.Builder, rendered []byte, idx int, total int) {
	if idx > 0 {
		builder.WriteString("\n")
	}
	builder.Write(rendered)
	if idx < total-1 {
		builder.WriteString("\n")
	}
}

func (r *MarkdownRenderer) renderEntryArticleMarkdown(entry model.Entry, grouped bool, redirectedFromURL string) []byte {
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
		fmt.Fprintf(&builder, mdFmtH1, entry.Headword)
	}
	if article.Dictionary != "" {
		fmt.Fprintf(&builder, mdFmtBlock, article.Dictionary)
	}
	if article.Edition != "" {
		fmt.Fprintf(&builder, mdFmtBlock, article.Edition)
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
	if citation := strings.TrimSpace(renderCitationMarkdown(article, redirectedFromURL)); citation != "" {
		builder.WriteString("\n\n")
		builder.WriteString(citation)
	}
	return []byte(strings.TrimSpace(builder.String()))
}

func (r *MarkdownRenderer) renderMarkdownSection(section model.Section, indent string) string {
	var builder strings.Builder
	heading := buildSectionHeading(section)

	if heading != "" {
		builder.WriteString(indent)
		builder.WriteString(heading)
	}

	firstParagraphInline := false
	for _, block := range sectionBlocks(section) {
		r.appendMarkdownBlock(&builder, block, heading, indent, &firstParagraphInline)
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

// buildSectionHeading assembles the display heading for a section from its Label and Title fields.
func buildSectionHeading(section model.Section) string {
	heading := strings.TrimSpace(section.Label)
	if title := strings.TrimSpace(section.Title); title != "" {
		if heading != "" {
			heading += " "
		}
		heading += title
	}
	return heading
}

// appendMarkdownBlock writes a single block's rendered text into builder.
// heading and firstParagraphInline control whether the first paragraph is inlined after the heading.
func (r *MarkdownRenderer) appendMarkdownBlock(
	builder *strings.Builder,
	block model.Block,
	heading, indent string,
	firstParagraphInline *bool,
) {
	switch block.Kind {
	case model.ArticleBlockKindParagraph:
		r.appendMarkdownParagraphBlock(builder, block, heading, indent, firstParagraphInline)
	case model.ArticleBlockKindTable:
		appendMarkdownTableBlock(builder, block, indent)
	}
}

func (r *MarkdownRenderer) appendMarkdownParagraphBlock(
	builder *strings.Builder,
	block model.Block,
	heading, indent string,
	firstParagraphInline *bool,
) {
	if block.Paragraph == nil {
		return
	}
	text := strings.TrimSpace(r.renderMarkdownParagraph(*block.Paragraph))
	if text == "" {
		return
	}
	if heading != "" && !*firstParagraphInline {
		builder.WriteString(" ")
		builder.WriteString(text)
		*firstParagraphInline = true
		return
	}
	if builder.Len() > 0 {
		builder.WriteString("\n\n")
	}
	builder.WriteString(indentLines(text, indent))
}

func appendMarkdownTableBlock(builder *strings.Builder, block model.Block, indent string) {
	if block.Table == nil {
		return
	}
	tableText := renderutil.RenderTableMarkdown(*block.Table, indent)
	if tableText == "" {
		return
	}
	if builder.Len() > 0 {
		builder.WriteString("\n\n")
	}
	builder.WriteString(tableText)
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
		return renderutil.RenderMarkdownInlines(paragraph.Inlines)
	}
	return normalizeLegacyMarkdownProjection(paragraph.Markdown)
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

func indentLines(text, indent string) string {
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

func renderCitationMarkdown(article *model.Article, redirectedFromURL string) string {
	if article == nil {
		return ""
	}

	citation := article.Citation
	if lines := structuredCitationLines(article, citation, redirectedFromURL); len(lines) > 0 {
		return strings.Join(lines, "\n")
	}
	if lines := fallbackCitationLines(article, citation); len(lines) > 0 {
		return strings.Join(lines, "\n")
	}
	return strings.TrimSpace(citation.Text)
}

func structuredCitationLines(article *model.Article, citation model.Citation, redirectedFromURL string) []string {
	lines := make([]string, 0, 5)
	if value := strings.TrimSpace(citation.SourceLabel); value != "" {
		lines = append(lines, "Source: "+value)
	}
	if len(lines) == 0 {
		return nil
	}
	if value := strings.TrimSpace(article.Dictionary); value != "" {
		lines = append(lines, "Dictionary: "+value)
	}
	if value := strings.TrimSpace(citation.Edition); value != "" {
		lines = append(lines, "Edition: "+value)
	}
	if value := citationURL(citation.CanonicalURL, redirectedFromURL); value != "" {
		lines = append(lines, "URL: "+value)
	}
	if value := strings.TrimSpace(citation.ConsultedAt); value != "" {
		lines = append(lines, "Consulted: "+value)
	}
	return lines
}

func fallbackCitationLines(article *model.Article, citation model.Citation) []string {
	lines := make([]string, 0, 1)
	if value := strings.TrimSpace(article.Dictionary); value != "" {
		lines = append(lines, "Dictionary: "+value)
	}
	if len(lines) == 0 || strings.TrimSpace(citation.Text) != "" {
		return nil
	}
	return lines
}

func citationURL(canonicalURL, redirectedFromURL string) string {
	value := strings.TrimSpace(canonicalURL)
	if value == "" {
		return ""
	}
	if redirectedFromURL == "" {
		return value
	}
	return redirectedFromURL + " → " + value
}
