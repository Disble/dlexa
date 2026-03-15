package renderutil

import (
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

func TestRenderTableMarkdownSimple(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{{Text: "Con tilde"}, {Text: "Sin tilde"}}}},
		Rows:    []model.TableRow{{Cells: []model.TableCell{{Text: "aún"}, {Text: "aun"}}}, {Cells: []model.TableCell{{Text: "sólo"}, {Text: "solo"}}}},
	}

	got := RenderTableMarkdown(table, "")
	for _, want := range []string{
		"| Con tilde | Sin tilde |",
		"|-----------|-----------|",
		"| aún       | aun       |",
		"| sólo      | solo      |",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderTableMarkdown() missing %q\n%s", want, got)
		}
	}
}

func TestRenderTableMarkdownWithIndent(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{{Text: "A"}, {Text: "B"}}}},
		Rows:    []model.TableRow{{Cells: []model.TableCell{{Text: "1"}, {Text: "2"}}}},
	}

	got := RenderTableMarkdown(table, "  ")
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "  ") {
			t.Fatalf("expected indent, got line: %q", line)
		}
	}
}

func TestRenderTableMarkdownEmpty(t *testing.T) {
	// An empty table (no headers, no rows) falls back to HTML rendering
	// because isSimpleMarkdownTable returns false (needs exactly 1 header row).
	table := model.Table{}
	got := RenderTableMarkdown(table, "")
	if !strings.Contains(got, "<table>") {
		t.Fatalf("RenderTableMarkdown(empty) = %q, want HTML fallback", got)
	}
}

func TestRenderTableMarkdownFallsBackToHTMLForComplexTable(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{{Text: "Título", ColSpan: 4}}}},
		Rows:    []model.TableRow{{Cells: []model.TableCell{{Text: "Con tilde", RowSpan: 2}, {Text: "Caso"}, {Text: "Ejemplo"}}}},
	}

	got := RenderTableMarkdown(table, "")
	if !strings.Contains(got, "<table>") {
		t.Fatalf("RenderTableMarkdown(complex) should fall back to HTML\n%s", got)
	}
	if strings.Contains(got, "|---") {
		t.Fatalf("RenderTableMarkdown(complex) should not have markdown divider\n%s", got)
	}
}

func TestRenderTableHTML(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{{Text: "A"}, {Text: "B"}}}},
		Rows:    []model.TableRow{{Cells: []model.TableCell{{Text: "1"}, {Text: "2"}}}},
	}

	got := RenderTableHTML(table, "")
	for _, want := range []string{
		"<table>",
		"<thead>",
		"<th>A</th>",
		"<th>B</th>",
		"</thead>",
		"<tbody>",
		"<td>1</td>",
		"<td>2</td>",
		"</tbody>",
		"</table>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderTableHTML() missing %q\n%s", want, got)
		}
	}
}

func TestRenderTableHTMLWithSpans(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{{Text: "Título", ColSpan: 4}}}},
		Rows:    []model.TableRow{{Cells: []model.TableCell{{Text: "Con tilde", RowSpan: 2}}}},
	}

	got := RenderTableHTML(table, "")
	if !strings.Contains(got, `colspan="4"`) {
		t.Fatalf("missing colspan\n%s", got)
	}
	if !strings.Contains(got, `rowspan="2"`) {
		t.Fatalf("missing rowspan\n%s", got)
	}
}

func TestRenderTableHTMLWithIndent(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{{Text: "A"}}}},
	}

	got := RenderTableHTML(table, "  ")
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "  ") {
			t.Fatalf("expected indent, got line: %q", line)
		}
	}
}

func TestRenderTableHTMLCellContentWithEmphasis(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{{Text: "*qué* / que"}}}},
	}

	got := RenderTableHTML(table, "")
	if !strings.Contains(got, "<em>qué</em> / que") {
		t.Fatalf("HTML cell content should convert markdown emphasis\n%s", got)
	}
}

func TestRenderTableHTMLCellContentWithInlines(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{
			{Text: "*qué* / que", Inlines: []model.Inline{{Kind: model.InlineKindEmphasis, Text: "qué"}, {Kind: model.InlineKindText, Text: " / que"}}},
		}}},
	}

	got := RenderTableHTML(table, "")
	if !strings.Contains(got, "<em>qué</em> / que") {
		t.Fatalf("HTML cell with inlines should use inline rendering\n%s", got)
	}
}

func TestRenderTableMarkdownWithInlineCells(t *testing.T) {
	table := model.Table{
		Headers: []model.TableRow{{Cells: []model.TableCell{
			{Text: "*Con* tilde", Inlines: []model.Inline{{Kind: model.InlineKindMention, Text: "Con"}, {Kind: model.InlineKindText, Text: " tilde"}}},
			{Text: "Sin tilde"},
		}}},
		Rows: []model.TableRow{{Cells: []model.TableCell{
			{Text: "*sólo*", Inlines: []model.Inline{{Kind: model.InlineKindMention, Text: "sólo"}}},
			{Text: "solo"},
		}}},
	}

	got := RenderTableMarkdown(table, "")
	if !strings.Contains(got, "| *Con* tilde |") {
		t.Fatalf("missing header inline\n%s", got)
	}
	if !strings.Contains(got, "| *sólo*") {
		t.Fatalf("missing row inline\n%s", got)
	}
}

func TestRenderHTMLFromMarkdownSubset(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"plain text", "hello world", "hello world"},
		{"emphasis", "*bold*", "<em>bold</em>"},
		{"newline", "line1\nline2", "line1<br>line2"},
		{"html escaping", "a < b & c > d", "a &lt; b &amp; c &gt; d"},
		{"carriage return normalization", "a\r\nb", "a<br>b"},
		{"bare carriage return", "a\rb", "a<br>b"},
		{"unclosed emphasis", "*incomplete", "*incomplete"},
		{"multiple emphasis", "*one* and *two*", "<em>one</em> and <em>two</em>"},
		{"emphasis with special chars", "*café*", "<em>café</em>"},
		{"text after emphasis", "before *bold* after", "before <em>bold</em> after"},
		{"only newlines", "\n\n", "<br><br>"},
		{"reference in text", "see → [link](http://example.com) here", `see &rarr; <a href="http://example.com"> [link</a> here`},
		{"unclosed reference bracket", "→ [incomplete", "→ [incomplete"},
		{"reference without close paren", "→ [label](url", "→ [label](url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderHTMLFromMarkdownSubset(tt.raw)
			if got != tt.want {
				t.Fatalf("RenderHTMLFromMarkdownSubset(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNormalizeMarkdownTableCellText(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"simple text", "hello", "hello"},
		{"with newlines", "line1\nline2", "line1<br>line2"},
		{"with pipes", "a|b", "a\\|b"},
		{"with crlf", "a\r\nb", "a<br>b"},
		{"with cr", "a\rb", "a<br>b"},
		{"with surrounding spaces", "  hello  ", "hello"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeMarkdownTableCellText(tt.raw)
			if got != tt.want {
				t.Fatalf("NormalizeMarkdownTableCellText(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
