package normalize

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/parse"
)

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

func NewDPDNormalizer() *DPDNormalizer {
	return &DPDNormalizer{}
}

func (n *DPDNormalizer) Normalize(ctx context.Context, descriptor model.SourceDescriptor, result parse.Result) ([]model.Entry, []model.Warning, error) {
	_ = ctx
	entries := make([]model.Entry, 0, len(result.Articles))
	warnings := []model.Warning{accessProfileWarning(descriptor.Name)}

	for _, article := range result.Articles {
		if len(article.Sections) == 0 {
			return nil, warnings, model.NewProblemError(model.Problem{
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

	return entries, warnings, nil
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
			if block.Paragraph == nil {
				continue
			}
			paragraph := normalizeParagraph(*block.Paragraph)
			if strings.TrimSpace(paragraph.Markdown) == "" && len(paragraph.Inlines) == 0 {
				continue
			}
			paragraphs = append(paragraphs, paragraph)
			paragraphCopy := paragraph
			normalizedBlocks = append(normalizedBlocks, model.Block{
				Kind:      model.ArticleBlockKindParagraph,
				Paragraph: &paragraphCopy,
			})
		case parse.ParsedBlockKindTable:
			if block.Table == nil {
				continue
			}
			table := normalizeTable(*block.Table)
			if len(table.Headers) == 0 && len(table.Rows) == 0 {
				continue
			}
			tableCopy := table
			normalizedBlocks = append(normalizedBlocks, model.Block{
				Kind:  model.ArticleBlockKindTable,
				Table: &tableCopy,
			})
		}
	}

	return normalizedBlocks, paragraphs
}

func normalizeParagraph(paragraph parse.ParsedParagraph) model.Paragraph {
	normalizedInlines := normalizeInlines(paragraph.Inlines)
	text := strings.TrimSpace(normalizeParagraphMarkdown(paragraph.HTML))
	if len(normalizedInlines) > 0 {
		text = strings.TrimSpace(renderInlineMarkdown(normalizedInlines))
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

func normalizeTableRows(input []parse.ParsedTableRow) []model.TableRow {
	rows := make([]model.TableRow, 0, len(input))
	for _, row := range input {
		cells := make([]model.TableCell, 0, len(row.Cells))
		lastNonEmpty := -1
		for idx, cell := range row.Cells {
			normalizedInlines := normalizeInlines(cell.Inlines)
			text := strings.TrimSpace(normalizeParagraphMarkdown(cell.HTML))
			if len(normalizedInlines) > 0 {
				text = strings.TrimSpace(renderInlineMarkdown(normalizedInlines))
			}
			if !hasSemanticInlines(normalizedInlines) {
				normalizedInlines = nil
			}
			cells = append(cells, model.TableCell{Text: text, Inlines: normalizedInlines, ColSpan: cell.ColSpan, RowSpan: cell.RowSpan})
			if text != "" {
				lastNonEmpty = idx
			}
		}
		if lastNonEmpty >= 0 {
			cells = cells[:lastNonEmpty+1]
		}
		if len(cells) == 0 {
			continue
		}
		rows = append(rows, model.TableRow{Cells: cells})
	}
	return rows
}

func normalizeInlines(input []model.Inline) []model.Inline {
	if len(input) == 0 {
		return nil
	}
	normalized := make([]model.Inline, 0, len(input))
	for _, inline := range input {
		text := cleanInlineText(inline.Text)
		if inline.Kind == model.InlineKindText {
			text = inline.Text
		}
		item := model.Inline{
			Kind:    inline.Kind,
			Variant: inline.Variant,
			Text:    text,
			Target:  inline.Target,
		}
		item.Children = normalizeInlines(inline.Children)
		switch item.Kind {
		case model.InlineKindReference:
			item.Text = strings.Trim(item.Text, "[] ")
			item.Text = strings.TrimPrefix(item.Text, "→")
			item.Text = cleanInlineText(item.Text)
		case model.InlineKindEmphasis:
			if len(item.Children) == 1 {
				child := item.Children[0]
				if child.Kind == model.InlineKindMention || child.Kind == model.InlineKindCorrection {
					item = child
				}
			}
		case model.InlineKindCitationQuote, model.InlineKindBibliography, model.InlineKindExample, model.InlineKindMention,
			model.InlineKindGloss, model.InlineKindWorkTitle, model.InlineKindSmallCaps, model.InlineKindEditorial,
			model.InlineKindPattern, model.InlineKindCorrection, model.InlineKindScaffold:
			if len(item.Children) > 0 {
				item.Text = cleanInlineText(renderInlineMarkdown(item.Children))
			}
		}
		if item.Kind == model.InlineKindText && strings.TrimSpace(item.Text) == "" {
			continue
		}
		normalized = append(normalized, item)
	}
	return cleanupReferenceSpacing(mergeNormalizedTextInlines(normalized))
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

func renderInlineMarkdown(inlines []model.Inline) string {
	var builder strings.Builder
	for _, inline := range inlines {
		piece := renderInlineMarkdownItem(inline)
		if builder.Len() > 0 && needsInlineSpace(builder.String(), piece) {
			builder.WriteString(" ")
		}
		builder.WriteString(piece)
	}
	return strings.TrimSpace(builder.String())
}

func renderInlineMarkdownItem(inline model.Inline) string {
	children := renderInlineMarkdown(inline.Children)
	text := inline.Text
	if children != "" {
		text = children
	}
	switch inline.Kind {
	case model.InlineKindExample:
		return "‹" + text + "›"
	case model.InlineKindMention, model.InlineKindEmphasis, model.InlineKindWorkTitle, model.InlineKindCorrection:
		if len(inline.Children) > 0 {
			return renderStyledInlineMarkdown(inline.Children, "*")
		}
		return "*" + text + "*"
	case model.InlineKindReference:
		if text == "" {
			return ""
		}
		return "→ [" + text + "](" + inline.Target + ")"
	case model.InlineKindScaffold:
		return text
	case model.InlineKindCitationQuote:
		return "«" + text + "»"
	default:
		return text
	}
}

func renderStyledInlineMarkdown(children []model.Inline, marker string) string {
	var builder strings.Builder
	appendPiece := func(piece string) {
		if piece == "" {
			return
		}
		if builder.Len() > 0 && !shouldGlueInlineWordBoundary(builder.String(), piece) && needsInlineSpace(builder.String(), piece) {
			builder.WriteString(" ")
		}
		builder.WriteString(piece)
	}
	flush := func(buffer []model.Inline) {
		text := strings.TrimSpace(renderInlineMarkdown(buffer))
		if text == "" {
			return
		}
		piece := text
		if shouldWrapStyledBuffer(buffer) {
			piece = marker + text + marker
		}
		appendPiece(piece)
	}

	buffer := make([]model.Inline, 0, len(children))
	for _, child := range children {
		if child.Kind != model.InlineKindScaffold {
			buffer = append(buffer, child)
			continue
		}
		flush(buffer)
		buffer = buffer[:0]
		plain := strings.TrimSpace(renderInlineMarkdown(child.Children))
		if plain == "" {
			plain = strings.TrimSpace(child.Text)
		}
		if plain == "" {
			continue
		}
		appendPiece(plain)
	}
	flush(buffer)
	return strings.TrimSpace(builder.String())
}

func needsInlineSpace(current, next string) bool {
	if current == "" || next == "" {
		return false
	}
	last, _ := utf8.DecodeLastRuneInString(current)
	first, _ := utf8.DecodeRuneInString(next)
	if strings.ContainsRune(" [{«", first) || strings.ContainsRune(")]}.;,:!?»", first) {
		return false
	}
	if strings.ContainsRune(" ([{«", last) {
		return false
	}
	if unicode.IsSpace(last) || unicode.IsSpace(first) {
		return false
	}
	return true
}

func shouldGlueInlineWordBoundary(current, next string) bool {
	last, ok := lastInlineWordRune(current)
	if !ok {
		return false
	}
	first, ok := firstInlineWordRune(next)
	if !ok {
		return false
	}
	return unicode.IsLetter(last) && unicode.IsLetter(first)
}

func shouldWrapStyledBuffer(buffer []model.Inline) bool {
	if len(buffer) != 1 {
		return true
	}
	switch buffer[0].Kind {
	case model.InlineKindMention, model.InlineKindEmphasis, model.InlineKindWorkTitle, model.InlineKindCorrection, model.InlineKindExample:
		return false
	default:
		return true
	}
}

func lastInlineWordRune(raw string) (rune, bool) {
	trimmed := strings.TrimRightFunc(raw, unicode.IsSpace)
	for trimmed != "" {
		r, size := utf8.DecodeLastRuneInString(trimmed)
		if r == utf8.RuneError && size == 0 {
			return 0, false
		}
		if strings.ContainsRune("*_`~", r) {
			trimmed = trimmed[:len(trimmed)-size]
			continue
		}
		if strings.ContainsRune("])}>", r) {
			return 0, false
		}
		return r, true
	}
	return 0, false
}

func firstInlineWordRune(raw string) (rune, bool) {
	trimmed := strings.TrimLeftFunc(raw, unicode.IsSpace)
	for trimmed != "" {
		r, size := utf8.DecodeRuneInString(trimmed)
		if r == utf8.RuneError && size == 0 {
			return 0, false
		}
		if strings.ContainsRune("*_`~", r) {
			trimmed = trimmed[size:]
			continue
		}
		if strings.ContainsRune("[(<{→", r) {
			return 0, false
		}
		return r, true
	}
	return 0, false
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
	var parts []string
	for _, section := range article.Sections {
		parts = append(parts, renderSectionMarkdown(section))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func renderSectionMarkdown(section model.Section) string {
	var builder strings.Builder
	heading := strings.TrimSpace(section.Label)
	if title := strings.TrimSpace(section.Title); title != "" {
		if heading != "" {
			heading += " "
		}
		heading += title
	}
	if heading != "" {
		builder.WriteString(heading)
	}
	firstParagraphInline := false
	for _, block := range normalizeSectionBlocks(section) {
		switch block.Kind {
		case model.ArticleBlockKindParagraph:
			if block.Paragraph == nil {
				continue
			}
			text := strings.TrimSpace(block.Paragraph.Markdown)
			if len(block.Paragraph.Inlines) > 0 {
				text = strings.TrimSpace(renderInlineMarkdown(block.Paragraph.Inlines))
			}
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
			builder.WriteString(text)
			firstParagraphInline = true
		case model.ArticleBlockKindTable:
			if block.Table == nil {
				continue
			}
			if builder.Len() > 0 {
				builder.WriteString("\n\n")
			}
			builder.WriteString(renderTableMarkdown(*block.Table, ""))
			firstParagraphInline = true
		}
	}

	for _, child := range section.Children {
		childMarkdown := strings.TrimSpace(renderSectionMarkdown(child))
		if childMarkdown == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(childMarkdown)
	}

	return strings.TrimSpace(builder.String())
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

func renderTableMarkdown(table model.Table, indent string) string {
	if !isSimpleMarkdownTable(table) {
		return renderTableHTML(table, indent)
	}

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
		text := normalizeMarkdownTableCellText(cell.Text)
		if len(cell.Inlines) > 0 {
			text = normalizeMarkdownTableCellText(renderInlineMarkdown(cell.Inlines))
		}
		result = append(result, text)
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
	return "|" + strings.Join(parts, "|") + "|"
}

func isSimpleMarkdownTable(table model.Table) bool {
	if len(table.Headers) != 1 {
		return false
	}
	columnCount := len(table.Headers[0].Cells)
	if columnCount == 0 {
		return false
	}
	for _, row := range table.Headers {
		if !isSimpleMarkdownRow(row, columnCount) {
			return false
		}
	}
	for _, row := range table.Rows {
		if !isSimpleMarkdownRow(row, columnCount) {
			return false
		}
	}
	return true
}

func isSimpleMarkdownRow(row model.TableRow, columnCount int) bool {
	if len(row.Cells) != columnCount {
		return false
	}
	for _, cell := range row.Cells {
		if cell.ColSpan > 1 || cell.RowSpan > 1 {
			return false
		}
	}
	return true
}

func normalizeMarkdownTableCellText(raw string) string {
	text := strings.TrimSpace(raw)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\n", "<br>")
	text = strings.ReplaceAll(text, "|", "\\|")
	return text
}

func renderTableHTML(table model.Table, indent string) string {
	var builder strings.Builder
	write := func(line string) {
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(indent)
		builder.WriteString(line)
	}

	write("<table>")
	if len(table.Headers) > 0 {
		write("  <thead>")
		for _, row := range table.Headers {
			write("    <tr>")
			for _, cell := range row.Cells {
				write("      " + formatHTMLTableCell("th", cell))
			}
			write("    </tr>")
		}
		write("  </thead>")
	}
	if len(table.Rows) > 0 {
		write("  <tbody>")
		for _, row := range table.Rows {
			write("    <tr>")
			for _, cell := range row.Cells {
				write("      " + formatHTMLTableCell("td", cell))
			}
			write("    </tr>")
		}
		write("  </tbody>")
	}
	write("</table>")
	return builder.String()
}

func formatHTMLTableCell(tag string, cell model.TableCell) string {
	attrs := make([]string, 0, 2)
	if cell.ColSpan > 1 {
		attrs = append(attrs, fmt.Sprintf(" colspan=\"%d\"", cell.ColSpan))
	}
	if cell.RowSpan > 1 {
		attrs = append(attrs, fmt.Sprintf(" rowspan=\"%d\"", cell.RowSpan))
	}
	return fmt.Sprintf("<%s%s>%s</%s>", tag, strings.Join(attrs, ""), renderHTMLTableCellContent(cell), tag)
}

func renderHTMLTableCellContent(cell model.TableCell) string {
	text := strings.TrimSpace(cell.Text)
	if len(cell.Inlines) > 0 {
		text = strings.TrimSpace(renderInlineMarkdown(cell.Inlines))
	}
	return renderHTMLFromMarkdownSubset(text)
}

func renderHTMLFromMarkdownSubset(raw string) string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	var builder strings.Builder
	for i := 0; i < len(raw); {
		if strings.HasPrefix(raw[i:], "→ [") {
			closeLabel := strings.Index(raw[i+3:], "](")
			if closeLabel >= 0 {
				labelEnd := i + 3 + closeLabel
				closeTarget := strings.Index(raw[labelEnd+2:], ")")
				if closeTarget >= 0 {
					targetEnd := labelEnd + 2 + closeTarget
					label := raw[i+3 : labelEnd]
					target := raw[labelEnd+2 : targetEnd]
					builder.WriteString("&rarr; <a href=\"")
					builder.WriteString(html.EscapeString(target))
					builder.WriteString("\">")
					builder.WriteString(html.EscapeString(label))
					builder.WriteString("</a>")
					i = targetEnd + 1
					continue
				}
			}
		}
		if raw[i] == '*' {
			close := strings.IndexByte(raw[i+1:], '*')
			if close >= 0 {
				content := raw[i+1 : i+1+close]
				builder.WriteString("<em>")
				builder.WriteString(html.EscapeString(content))
				builder.WriteString("</em>")
				i += close + 2
				continue
			}
		}
		if raw[i] == '\n' {
			builder.WriteString("<br>")
			i++
			continue
		}
		next := i + 1
		for next < len(raw) && raw[next] != '*' && raw[next] != '\n' && !strings.HasPrefix(raw[next:], "→ [") {
			next++
		}
		builder.WriteString(html.EscapeString(raw[i:next]))
		i = next
	}
	return builder.String()
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
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func accessProfileWarning(source string) model.Warning {
	return model.Warning{
		Code:    "dpd_access_profile",
		Source:  source,
		Message: "validated access method: direct GET /dpd/<term> with browser-like User-Agent reaches article HTML; low-profile/no-UA requests may trigger Cloudflare challenge pages; /srv/keys is not useful here; go-rae is not a direct DPD blueprint",
	}
}
