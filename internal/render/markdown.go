package render

import (
	"context"
	"fmt"
	"html"
	"strings"
	"unicode"
	"unicode/utf8"

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
		if inline.Kind == model.InlineKindEmphasis && len(inline.Children) == 1 {
			child := inline.Children[0]
			if child.Kind == model.InlineKindMention || child.Kind == model.InlineKindCorrection {
				return renderMarkdownInline(child)
			}
		}
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
		if len(inline.Children) > 0 {
			return renderStyledMarkdownInline(inline.Children, "*")
		}
		return "*" + text + "*"
	case model.InlineKindReference:
		return "→ [" + text + "](" + inline.Target + ")"
	case model.InlineKindScaffold:
		return text
	default:
		return text
	}
}

func renderStyledMarkdownInline(children []model.Inline, marker string) string {
	var builder strings.Builder
	buffer := make([]model.Inline, 0, len(children))
	appendPiece := func(piece string) {
		if piece == "" {
			return
		}
		if builder.Len() > 0 && !shouldGlueInlineWordBoundary(builder.String(), piece) && needsInlineSpace(builder.String(), piece) {
			builder.WriteString(" ")
		}
		builder.WriteString(piece)
	}
	flush := func() {
		if len(buffer) == 0 {
			return
		}
		snapshot := append([]model.Inline(nil), buffer...)
		text := strings.TrimSpace(renderMarkdownInlines(buffer))
		buffer = buffer[:0]
		if text == "" {
			return
		}
		piece := text
		if shouldWrapStyledBuffer(snapshot) {
			piece = marker + text + marker
		}
		appendPiece(piece)
	}

	for _, child := range children {
		if child.Kind != model.InlineKindScaffold {
			buffer = append(buffer, child)
			continue
		}
		flush()
		piece := strings.TrimSpace(renderMarkdownInlines(child.Children))
		if piece == "" {
			piece = strings.TrimSpace(child.Text)
		}
		if piece == "" {
			continue
		}
		appendPiece(piece)
	}
	flush()
	return strings.TrimSpace(builder.String())
}

func needsInlineSpace(current, next string) bool {
	if current == "" || next == "" {
		return false
	}
	last, _ := utf8.DecodeLastRuneInString(current)
	first, _ := utf8.DecodeRuneInString(next)
	if strings.ContainsRune(" [{«", first) || strings.ContainsRune(")]}.;,!?»:", first) {
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
			text = normalizeMarkdownTableCellText(renderMarkdownInlines(cell.Inlines))
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
		text = strings.TrimSpace(renderMarkdownInlines(cell.Inlines))
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
