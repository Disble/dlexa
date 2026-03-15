package renderutil

import (
	"fmt"
	"html"
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

// InlineRenderer is a function that renders a slice of inlines to a markdown string.
// This allows table rendering to be parameterized by the inline rendering variant
// (normalize uses ‹› for examples, render uses *).
type InlineRenderer func(inlines []model.Inline) string

// RenderTableMarkdown renders a model.Table as markdown (pipe table) or
// falls back to HTML for complex tables (colspan/rowspan/multi-header).
// It uses RenderMarkdownInlines as the default inline renderer.
func RenderTableMarkdown(table model.Table, indent string) string {
	return RenderTableMarkdownWith(table, indent, RenderMarkdownInlines)
}

// RenderTableMarkdownWith renders a model.Table as markdown using the given
// inline renderer for cell content.
func RenderTableMarkdownWith(table model.Table, indent string, renderInlines InlineRenderer) string {
	if !isSimpleMarkdownTable(table) {
		return RenderTableHTMLWith(table, indent, renderInlines)
	}

	rows := make([][]string, 0, len(table.Headers)+len(table.Rows))
	for _, row := range table.Headers {
		rows = append(rows, tableRowTexts(row, renderInlines))
	}
	for _, row := range table.Rows {
		rows = append(rows, tableRowTexts(row, renderInlines))
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

// RenderTableHTML renders a model.Table as an HTML table string.
// It uses RenderMarkdownInlines as the default inline renderer.
func RenderTableHTML(table model.Table, indent string) string {
	return RenderTableHTMLWith(table, indent, RenderMarkdownInlines)
}

// RenderTableHTMLWith renders a model.Table as an HTML table string using
// the given inline renderer for cell content.
func RenderTableHTMLWith(table model.Table, indent string, renderInlines InlineRenderer) string {
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
				write("      " + formatHTMLTableCell("th", cell, renderInlines))
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
				write("      " + formatHTMLTableCell("td", cell, renderInlines))
			}
			write("    </tr>")
		}
		write("  </tbody>")
	}
	write("</table>")
	return builder.String()
}

// RenderHTMLFromMarkdownSubset converts a limited markdown subset (emphasis, references,
// newlines) to HTML. Used for table cell content in HTML fallback tables.
func RenderHTMLFromMarkdownSubset(raw string) string {
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
			closeIdx := strings.IndexByte(raw[i+1:], '*')
			if closeIdx >= 0 {
				content := raw[i+1 : i+1+closeIdx]
				builder.WriteString("<em>")
				builder.WriteString(html.EscapeString(content))
				builder.WriteString("</em>")
				i += closeIdx + 2
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

// NormalizeMarkdownTableCellText normalizes text for markdown table cells:
// trims whitespace, normalizes line endings, escapes pipes.
func NormalizeMarkdownTableCellText(raw string) string {
	text := strings.TrimSpace(raw)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\n", "<br>")
	text = strings.ReplaceAll(text, "|", "\\|")
	return text
}

func tableRowTexts(row model.TableRow, renderInlines InlineRenderer) []string {
	result := make([]string, 0, len(row.Cells))
	for _, cell := range row.Cells {
		text := NormalizeMarkdownTableCellText(cell.Text)
		if len(cell.Inlines) > 0 {
			text = NormalizeMarkdownTableCellText(renderInlines(cell.Inlines))
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

func formatHTMLTableCell(tag string, cell model.TableCell, renderInlines InlineRenderer) string {
	attrs := make([]string, 0, 2)
	if cell.ColSpan > 1 {
		attrs = append(attrs, fmt.Sprintf(" colspan=\"%d\"", cell.ColSpan))
	}
	if cell.RowSpan > 1 {
		attrs = append(attrs, fmt.Sprintf(" rowspan=\"%d\"", cell.RowSpan))
	}
	return fmt.Sprintf("<%s%s>%s</%s>", tag, strings.Join(attrs, ""), renderHTMLTableCellContent(cell, renderInlines), tag)
}

func renderHTMLTableCellContent(cell model.TableCell, renderInlines InlineRenderer) string {
	text := strings.TrimSpace(cell.Text)
	if len(cell.Inlines) > 0 {
		text = strings.TrimSpace(renderInlines(cell.Inlines))
	}
	return RenderHTMLFromMarkdownSubset(text)
}
