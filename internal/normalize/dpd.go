package normalize

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
	"github.com/Disble/dlexa/internal/renderutil"
)

// DPDNormalizer transforms DPD parse results into canonical domain entries.
type DPDNormalizer struct{}

var (
	reNormalizeReference = regexp.MustCompile(`(?is)\(?\s*→\s*<a\b[^>]*href="([^"]+)"[^>]*>(.*?)</a>\s*\)?`)
	reNormalizeDFN       = regexp.MustCompile(`(?is)<dfn>(.*?)</dfn>`)
	reNormalizeEmphasis  = regexp.MustCompile(`(?is)<(?:em|i)\b[^>]*>(.*?)</(?:em|i)>`)
	reNormalizeExample   = regexp.MustCompile(`(?is)<span class="ejemplo">(.*?)</span>`)
	reNormalizeTags      = regexp.MustCompile(`(?is)<[^>]+>`)
	reCitationURL        = regexp.MustCompile(`https://\S+`)
	reCitationConsulted  = regexp.MustCompile(`\[Consulta:\s*([^\]]+)\]`)
	reCitationPrefix     = regexp.MustCompile(`^(.*?):\s*`)
)

// NewDPDNormalizer returns a new DPDNormalizer instance.
func NewDPDNormalizer() *DPDNormalizer {
	return &DPDNormalizer{}
}

// Normalize converts parsed DPD articles and structured misses into normalized results.
func (n *DPDNormalizer) Normalize(ctx context.Context, descriptor model.SourceDescriptor, result parse.Result) (Result, error) {
	_ = ctx
	if result.Miss != nil {
		return Result{
			Miss:     normalizeLookupMiss(descriptor, *result.Miss),
			Warnings: []model.Warning{accessProfileWarning(descriptor.Name)},
		}, nil
	}

	entries := make([]model.Entry, 0, len(result.Articles))
	warnings := []model.Warning{accessProfileWarning(descriptor.Name)}

	for _, article := range result.Articles {
		if len(article.Sections) == 0 {
			return Result{Warnings: warnings}, model.NewProblemError(model.Problem{
				Code:     model.ProblemCodeDPDTransformFailed,
				Message:  "DPD article has no sections to normalize",
				Source:   descriptor.Name,
				Severity: model.ProblemSeverityError,
			}, nil)
		}

		sections := normalizeSections(article.Sections)
		normalizedArticle := model.Article{
			Dictionary:   article.Dictionary,
			Edition:      article.Edition,
			Lemma:        article.Lemma,
			CanonicalURL: article.CanonicalURL,
			Sections:     sections,
			Citation:     normalizeCitation(article),
		}

		entryID := strings.TrimSpace(article.EntryID)
		if entryID == "" {
			entryID = slug(article.Lemma)
		}

		entry := model.Entry{
			ID:       fmt.Sprintf("%s:%s#%s", descriptor.Name, slug(article.Lemma), slug(entryID)),
			Headword: article.Lemma,
			Summary:  article.Dictionary,
			Content:  markdownBody(normalizedArticle),
			Source:   descriptor.Name,
			URL:      article.CanonicalURL,
			Article:  &normalizedArticle,
			Metadata: map[string]string{
				"normalized_by":  "dpd",
				"access_profile": "browser-like direct /dpd/<term>",
				"entry_id":       entryID,
			},
		}
		entries = append(entries, entry)
	}

	return Result{Entries: entries, Warnings: warnings}, nil
}

func normalizeLookupMiss(descriptor model.SourceDescriptor, miss parse.ParsedLookupMiss) *model.LookupMiss {
	normalized := &model.LookupMiss{
		Kind:       model.LookupMissKind(miss.Kind),
		Query:      strings.TrimSpace(miss.Query),
		NoticeText: strings.TrimSpace(miss.NoticeText),
		Source:     descriptor.Name,
	}
	if miss.RelatedEntry != nil {
		normalized.Suggestion = &model.LookupSuggestion{
			Kind:         string(model.LookupMissKindRelatedEntry),
			DisplayText:  strings.TrimSpace(miss.RelatedEntry.DisplayText),
			EntryID:      strings.TrimSpace(miss.RelatedEntry.EntryID),
			URL:          strings.TrimSpace(miss.RelatedEntry.Href),
			RawLabelHTML: strings.TrimSpace(miss.RelatedEntry.RawLabelHTML),
		}
		return normalized
	}
	if normalized.Kind == model.LookupMissKindGenericNotFound {
		query := normalized.Query
		normalized.NextAction = &model.LookupNextAction{
			Kind:    model.LookupNextActionKindSearch,
			Query:   query,
			Command: "dlexa search " + query,
		}
	}
	return normalized
}

func normalizeSections(input []parse.ParsedSection) []model.Section {
	sections := make([]model.Section, 0, len(input))
	for _, section := range input {
		blocks, paragraphs := normalizeBlocks(section.Blocks, section.Paragraphs)
		children := normalizeSections(section.Children)
		if len(children) == 0 {
			children = nil
		}
		sections = append(sections, model.Section{
			Label:      section.Label,
			Title:      section.Title,
			Blocks:     blocks,
			Paragraphs: paragraphs,
			Children:   children,
		})
	}

	return sections
}

// normalizeBlocks normalizes a slice of parsed blocks, falling back to paragraphs
// when the block slice is empty.
// Cognitive complexity: ~8 (one early-return branch + one for + one switch with two cases each guarded by a single if).
func normalizeBlocks(blocks []parse.ParsedBlock, fallbackParagraphs []parse.ParsedParagraph) ([]model.Block, []model.Paragraph) {
	if len(blocks) == 0 && len(fallbackParagraphs) > 0 {
		blocks = make([]parse.ParsedBlock, 0, len(fallbackParagraphs))
		for _, paragraph := range fallbackParagraphs {
			paragraphCopy := paragraph
			blocks = append(blocks, parse.ParsedBlock{Kind: parse.ParsedBlockKindParagraph, Paragraph: &paragraphCopy})
		}
	}

	normalizedBlocks := make([]model.Block, 0, len(blocks))
	paragraphs := make([]model.Paragraph, 0, len(blocks))
	for _, block := range blocks {
		switch block.Kind {
		case parse.ParsedBlockKindParagraph:
			b, p, ok := normalizeParsedBlockParagraph(block)
			if !ok {
				continue
			}
			paragraphs = append(paragraphs, p)
			normalizedBlocks = append(normalizedBlocks, b)
		case parse.ParsedBlockKindTable:
			b, ok := normalizeParsedBlockTable(block)
			if !ok {
				continue
			}
			normalizedBlocks = append(normalizedBlocks, b)
		}
	}

	return normalizedBlocks, paragraphs
}

// normalizeParsedBlockParagraph normalizes a paragraph block. Returns ok=false when
// the block should be skipped (nil pointer or empty content after normalization).
func normalizeParsedBlockParagraph(block parse.ParsedBlock) (model.Block, model.Paragraph, bool) {
	if block.Paragraph == nil {
		return model.Block{}, model.Paragraph{}, false
	}
	paragraph := normalizeParagraph(*block.Paragraph)
	if strings.TrimSpace(paragraph.Markdown) == "" && len(paragraph.Inlines) == 0 {
		return model.Block{}, model.Paragraph{}, false
	}
	paragraphCopy := paragraph
	b := model.Block{
		Kind:      model.ArticleBlockKindParagraph,
		Paragraph: &paragraphCopy,
	}
	return b, paragraph, true
}

// normalizeParsedBlockTable normalizes a table block. Returns ok=false when the
// block should be skipped (nil pointer or empty table after normalization).
func normalizeParsedBlockTable(block parse.ParsedBlock) (model.Block, bool) {
	if block.Table == nil {
		return model.Block{}, false
	}
	table := normalizeTable(*block.Table)
	if len(table.Headers) == 0 && len(table.Rows) == 0 {
		return model.Block{}, false
	}
	tableCopy := table
	return model.Block{
		Kind:  model.ArticleBlockKindTable,
		Table: &tableCopy,
	}, true
}

func normalizeParagraph(paragraph parse.ParsedParagraph) model.Paragraph {
	normalizedInlines := normalizeInlines(paragraph.Inlines)
	text := strings.TrimSpace(normalizeParagraphMarkdown(paragraph.HTML))
	if len(normalizedInlines) > 0 {
		text = strings.TrimSpace(renderutil.RenderInlineMarkdown(normalizedInlines))
	}
	if !hasSemanticInlines(normalizedInlines) {
		normalizedInlines = nil
	}
	return model.Paragraph{Markdown: text, Inlines: normalizedInlines}
}

func hasSemanticInlines(inlines []model.Inline) bool {
	for _, inline := range inlines {
		if inline.Kind != model.InlineKindText {
			return true
		}
		if hasSemanticInlines(inline.Children) {
			return true
		}
	}
	return false
}

func normalizeTable(input parse.ParsedTable) model.Table {
	return model.Table{
		Headers: normalizeTableRows(input.Headers),
		Rows:    normalizeTableRows(input.Rows),
	}
}

// normalizeTableRows normalizes a slice of table rows, resolving inline content
// and trimming trailing empty cells from each row.
// Cognitive complexity: ~8 (one outer for + one inner for + two if guards).
func normalizeTableRows(input []parse.ParsedTableRow) []model.TableRow {
	rows := make([]model.TableRow, 0, len(input))
	for _, row := range input {
		cells := normalizeTableRowCells(row.Cells)
		if len(cells) == 0 {
			continue
		}
		rows = append(rows, model.TableRow{Cells: cells})
	}
	return rows
}

// normalizeTableRowCells resolves inline content for every cell in a row and
// trims trailing empty cells. Returns nil when the trimmed result is empty.
func normalizeTableRowCells(rawCells []parse.ParsedTableCell) []model.TableCell {
	cells := make([]model.TableCell, 0, len(rawCells))
	lastNonEmpty := -1
	for idx, cell := range rawCells {
		text, inlines := resolveTableCellContent(cell)
		cells = append(cells, model.TableCell{Text: text, Inlines: inlines, ColSpan: cell.ColSpan, RowSpan: cell.RowSpan})
		if text != "" {
			lastNonEmpty = idx
		}
	}
	if lastNonEmpty < 0 {
		return nil
	}
	return cells[:lastNonEmpty+1]
}

// resolveTableCellContent computes the display text and semantic inlines for a
// single parsed table cell.
func resolveTableCellContent(cell parse.ParsedTableCell) (string, []model.Inline) {
	normalizedInlines := normalizeInlines(cell.Inlines)
	text := strings.TrimSpace(normalizeParagraphMarkdown(cell.HTML))
	if len(normalizedInlines) > 0 {
		text = strings.TrimSpace(renderutil.RenderInlineMarkdown(normalizedInlines))
	}
	if !hasSemanticInlines(normalizedInlines) {
		normalizedInlines = nil
	}
	return text, normalizedInlines
}

// normalizeInlines normalizes a slice of inline elements, applying kind-specific
// text transforms and merging adjacent plain-text spans.
// Cognitive complexity: ~9 (nil guard + one for + one if for empty text skip).
func normalizeInlines(input []model.Inline) []model.Inline {
	if len(input) == 0 {
		return nil
	}
	normalized := make([]model.Inline, 0, len(input))
	for _, inline := range input {
		item := buildNormalizedInline(inline)
		if item.Kind == model.InlineKindText && strings.TrimSpace(item.Text) == "" {
			continue
		}
		normalized = append(normalized, item)
	}
	return cleanupReferenceSpacing(mergeNormalizedTextInlines(normalized))
}

// buildNormalizedInline constructs a single normalized Inline from a raw parsed
// inline, applying all kind-specific text and child transforms.
// Cognitive complexity: ~7 (one switch with four cases, one nested if).
func buildNormalizedInline(inline model.Inline) model.Inline {
	text := cleanInlineText(inline.Text)
	if inline.Kind == model.InlineKindText {
		text = inline.Text
	}
	item := model.Inline{
		Kind:     inline.Kind,
		Variant:  inline.Variant,
		Text:     text,
		Target:   inline.Target,
		Children: normalizeInlines(inline.Children),
	}
	return applyInlineKindTransforms(item)
}

// applyInlineKindTransforms applies kind-specific post-processing to an already
// constructed inline item (text cleanup, child collapsing, etc.).
// Cognitive complexity: ~7 (one switch with three meaningful cases, one nested if).
func applyInlineKindTransforms(item model.Inline) model.Inline {
	switch item.Kind {
	case model.InlineKindReference:
		item.Text = strings.Trim(item.Text, "[] ")
		item.Text = strings.TrimPrefix(item.Text, "→")
		item.Text = cleanInlineText(item.Text)
	case model.InlineKindEmphasis:
		if len(item.Children) == 1 {
			child := item.Children[0]
			if child.Kind == model.InlineKindMention || child.Kind == model.InlineKindCorrection {
				return child
			}
		}
	case model.InlineKindCitationQuote, model.InlineKindBibliography, model.InlineKindExample, model.InlineKindMention,
		model.InlineKindGloss, model.InlineKindWorkTitle, model.InlineKindSmallCaps, model.InlineKindEditorial,
		model.InlineKindPattern, model.InlineKindCorrection, model.InlineKindScaffold:
		if len(item.Children) > 0 {
			item.Text = cleanInlineText(renderutil.RenderInlineMarkdown(item.Children))
		}
	}
	return item
}

func mergeNormalizedTextInlines(inlines []model.Inline) []model.Inline {
	merged := make([]model.Inline, 0, len(inlines))
	for _, inline := range inlines {
		if inline.Kind == model.InlineKindText && len(merged) > 0 && merged[len(merged)-1].Kind == model.InlineKindText {
			merged[len(merged)-1].Text = merged[len(merged)-1].Text + inline.Text
			continue
		}
		merged = append(merged, inline)
	}
	return merged
}

func cleanupReferenceSpacing(inlines []model.Inline) []model.Inline {
	for idx := range inlines {
		if inlines[idx].Kind != model.InlineKindReference || idx == 0 {
			continue
		}
		prev := &inlines[idx-1]
		if prev.Kind != model.InlineKindText {
			continue
		}
		prev.Text = strings.ReplaceAll(prev.Text, "→", "")
		prev.Text = strings.TrimRight(prev.Text, " ")
	}
	filtered := inlines[:0]
	for _, inline := range inlines {
		if inline.Kind == model.InlineKindText && strings.TrimSpace(inline.Text) == "" {
			continue
		}
		filtered = append(filtered, inline)
	}
	return filtered
}

func normalizeParagraphMarkdown(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	raw = reNormalizeReference.ReplaceAllStringFunc(raw, func(match string) string {
		parts := reNormalizeReference.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		label := cleanInlineText(parts[2])
		label = strings.TrimPrefix(label, "→")
		label = strings.Trim(label, "[] ")
		reference := fmt.Sprintf("→ [%s](%s)", label, html.UnescapeString(parts[1]))
		trimmed := strings.TrimSpace(match)
		if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") {
			return "(" + reference + ")"
		}
		return reference
	})
	raw = reNormalizeDFN.ReplaceAllStringFunc(raw, func(match string) string {
		parts := reNormalizeDFN.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		return cleanInlineText(parts[1])
	})
	raw = reNormalizeExample.ReplaceAllStringFunc(raw, func(match string) string {
		parts := reNormalizeExample.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		return "‹" + cleanInlineText(parts[1]) + "›"
	})
	raw = reNormalizeEmphasis.ReplaceAllStringFunc(raw, func(match string) string {
		parts := reNormalizeEmphasis.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		return "*" + cleanInlineText(parts[1]) + "*"
	})
	raw = reNormalizeTags.ReplaceAllString(raw, "")
	return cleanInlineText(raw)
}

func normalizeCitation(article parse.ParsedArticle) model.Citation {
	text := cleanInlineText(article.Citation.Text)
	citation := model.Citation{
		SourceLabel:  article.Dictionary,
		CanonicalURL: article.CanonicalURL,
		Edition:      article.Edition,
		Text:         text,
	}

	if prefix := reCitationPrefix.FindStringSubmatch(text); len(prefix) == 2 {
		citation.SourceLabel = cleanInlineText(prefix[1])
	}
	if url := reCitationURL.FindString(text); url != "" {
		citation.CanonicalURL = strings.TrimRight(url, ",.;)")
	}
	if consulted := reCitationConsulted.FindStringSubmatch(text); len(consulted) == 2 {
		citation.ConsultedAt = cleanInlineText(consulted[1])
	}

	return citation
}

func markdownBody(article model.Article) string {
	parts := make([]string, 0, len(article.Sections))
	for _, section := range article.Sections {
		parts = append(parts, renderSectionMarkdown(section))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

// renderSectionMarkdown renders a section and its children into a markdown string.
// Cognitive complexity: ~8 (one for + one for-children + two if guards).
func renderSectionMarkdown(section model.Section) string {
	var builder strings.Builder
	heading := buildSectionHeading(section)
	if heading != "" {
		builder.WriteString(heading)
	}
	firstParagraphInline := false
	for _, block := range normalizeSectionBlocks(section) {
		written := renderBlockMarkdown(&builder, block, heading, firstParagraphInline)
		if written {
			firstParagraphInline = true
		}
	}
	appendChildSections(&builder, section.Children)
	return strings.TrimSpace(builder.String())
}

// buildSectionHeading assembles the heading string from label and optional title.
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

// renderBlockMarkdown writes one block to the builder.
// Returns true when content was actually written.
// Cognitive complexity: ~8 (one switch with two cases, each guarded by an if).
func renderBlockMarkdown(builder *strings.Builder, block model.Block, heading string, firstParagraphInline bool) bool {
	switch block.Kind {
	case model.ArticleBlockKindParagraph:
		return renderParagraphBlock(builder, block, heading, firstParagraphInline)
	case model.ArticleBlockKindTable:
		return renderTableBlock(builder, block)
	}
	return false
}

// renderParagraphBlock renders a paragraph block into the builder.
// Returns true when content was written.
func renderParagraphBlock(builder *strings.Builder, block model.Block, heading string, firstParagraphInline bool) bool {
	if block.Paragraph == nil {
		return false
	}
	text := strings.TrimSpace(block.Paragraph.Markdown)
	if len(block.Paragraph.Inlines) > 0 {
		text = strings.TrimSpace(renderutil.RenderInlineMarkdown(block.Paragraph.Inlines))
	}
	if text == "" {
		return false
	}
	if heading != "" && !firstParagraphInline {
		builder.WriteString(" ")
		builder.WriteString(text)
		return true
	}
	if builder.Len() > 0 {
		builder.WriteString("\n\n")
	}
	builder.WriteString(text)
	return true
}

// renderTableBlock renders a table block into the builder.
// Returns true when content was written.
func renderTableBlock(builder *strings.Builder, block model.Block) bool {
	if block.Table == nil {
		return false
	}
	if builder.Len() > 0 {
		builder.WriteString("\n\n")
	}
	builder.WriteString(renderutil.RenderTableMarkdownWith(*block.Table, "", renderutil.RenderInlineMarkdown))
	return true
}

// appendChildSections renders each non-empty child section and appends it to the builder.
func appendChildSections(builder *strings.Builder, children []model.Section) {
	for _, child := range children {
		childMarkdown := strings.TrimSpace(renderSectionMarkdown(child))
		if childMarkdown == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(childMarkdown)
	}
}

func normalizeSectionBlocks(section model.Section) []model.Block {
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

func slug(raw string) string {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	cleaned = strings.Join(strings.Fields(cleaned), "-")
	if cleaned == "" {
		return "entry"
	}
	return cleaned
}

func cleanInlineText(raw string) string {
	text := html.UnescapeString(raw)
	text = strings.ReplaceAll(text, "\u200d", "")

	// Preserve validated and speculative DPD signs before whitespace collapsing.
	// WARNING: *, ‖, and // are speculative only; no real DPD HTML validation
	// exists yet, so these inferred patterns MUST be revisited when examples are found.
	preservedSigns := []string{
		"\u2297", // ⊗ exclusion
		"@",      // digital edition
		"+",      // construction marker
		"*",      // agrammatical (SPECULATIVE)
		"‖",      // hypothetical (SPECULATIVE)
		"//",     // phoneme (SPECULATIVE)
	}
	placeholders := make([]string, len(preservedSigns))
	for idx, sign := range preservedSigns {
		placeholder := fmt.Sprintf("\x00DPDSIGN%d\x00", idx)
		placeholders[idx] = placeholder
		text = strings.ReplaceAll(text, sign, placeholder)
	}

	text = strings.Join(strings.Fields(text), " ")
	for idx, placeholder := range placeholders {
		text = strings.ReplaceAll(text, placeholder, preservedSigns[idx])
	}

	return strings.TrimSpace(text)
}

func accessProfileWarning(source string) model.Warning {
	return model.Warning{
		Code:    "dpd_access_profile",
		Source:  source,
		Message: "validated access method: direct GET /dpd/<term> with browser-like User-Agent reaches article HTML; low-profile/no-UA requests may trigger Cloudflare challenge pages; /srv/keys is useful for entry discovery only; search normalization owns conservative label interpretation and must stay format-neutral instead of emitting renderer-specific projections; go-rae is not a direct DPD blueprint",
	}
}
