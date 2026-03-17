package normalize

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

const (
	dpdDictionary   = "Diccionario panhispánico de dudas"
	dpdEdition      = "2.ª edición"
	dpdNormalizeErr = "Normalize() error = %v"
	dpdDespues      = "Después"
	dpdTildeURL     = "https://www.rae.es/dpd/tilde"
	dpdConTilde     = "Con tilde"
)

func normalizeSingleParagraph(t *testing.T, raw string) string {
	t.Helper()
	return normalizeParagraphMarkdown(raw)
}

func TestDPDNormalizerBuildsMarkdownReadyArticle(t *testing.T) {
	normalizer := NewDPDNormalizer()
	result := parse.Result{Articles: []parse.ParsedArticle{{
		Dictionary:   dpdDictionary,
		Edition:      dpdEdition,
		EntryID:      "bien",
		Lemma:        "bien",
		CanonicalURL: "https://www.rae.es/dpd/bien",
		Sections: []parse.ParsedSection{
			{Label: "1.", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Uno"}}}, Paragraphs: []parse.ParsedParagraph{{HTML: "Uno"}}},
			{Label: "6.", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Seis"}}}, Paragraphs: []parse.ParsedParagraph{{HTML: "Seis"}}, Children: []parse.ParsedSection{{Label: "a)", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Hijo A"}}}, Paragraphs: []parse.ParsedParagraph{{HTML: "Hijo A"}}}, {Label: "b)", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Hijo B"}}}, Paragraphs: []parse.ParsedParagraph{{HTML: "Hijo B"}}}, {Label: "c)", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Hijo C"}}}, Paragraphs: []parse.ParsedParagraph{{HTML: "Hijo C"}}}}},
		},
		Citation: parse.ParsedCitation{Text: "Real Academia Española..."},
	}}}

	entries, warnings, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, result)
	if err != nil {
		t.Fatalf(dpdNormalizeErr, err)
	}
	if len(warnings) == 0 || warnings[0].Code != "dpd_access_profile" {
		t.Fatalf("warnings = %#v, want access profile warning", warnings)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}

	entry := entries[0]
	if entry.Article == nil {
		t.Fatal("entry.Article = nil")
	}
	if got := len(entry.Article.Sections); got != 2 {
		t.Fatalf("sections = %d, want 2", got)
	}
	if got := len(entry.Article.Sections[1].Children); got != 3 {
		t.Fatalf("nested children = %d, want 3", got)
	}
	if entry.Metadata["access_profile"] != "browser-like direct /dpd/<term>" {
		t.Fatalf("access_profile = %q", entry.Metadata["access_profile"])
	}
	if entry.ID != "dpd:bien#bien" {
		t.Fatalf("entry id = %q", entry.ID)
	}
	if want := "1. Uno\n\n6. Seis\n\na) Hijo A\n\nb) Hijo B\n\nc) Hijo C"; entry.Content != want {
		t.Fatalf("content = %q, want %q", entry.Content, want)
	}
}

func TestDPDNormalizerPreservesMixedBlockOrderingAndParagraphProjection(t *testing.T) {
	sections := normalizeSections([]parse.ParsedSection{{
		Label: "1.",
		Blocks: []parse.ParsedBlock{
			{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Antes"}},
			{Kind: parse.ParsedBlockKindTable, Table: &parse.ParsedTable{Headers: []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: "A"}, {HTML: "B"}}}}, Rows: []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: "1"}, {HTML: "2"}}}}}},
			{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: dpdDespues}},
		},
	}})

	section := sections[0]
	if len(section.Blocks) != 3 {
		t.Fatalf("blocks = %d, want 3", len(section.Blocks))
	}
	if section.Blocks[0].Kind != model.ArticleBlockKindParagraph || section.Blocks[1].Kind != model.ArticleBlockKindTable || section.Blocks[2].Kind != model.ArticleBlockKindParagraph {
		t.Fatalf("block kinds = %#v", section.Blocks)
	}
	if len(section.Paragraphs) != 2 {
		t.Fatalf("paragraph projection = %d, want 2", len(section.Paragraphs))
	}
	if section.Blocks[1].Table == nil || section.Blocks[1].Table.Rows[0].Cells[1].Text != "2" {
		t.Fatalf("table = %#v", section.Blocks[1].Table)
	}
}

func TestDPDNormalizerBuildsDistinctIDsForDuplicateLemmas(t *testing.T) {
	result := parse.Result{Articles: []parse.ParsedArticle{
		{Dictionary: dpdDictionary, Edition: dpdEdition, EntryID: "tilde", Lemma: "tilde", CanonicalURL: dpdTildeURL, Sections: []parse.ParsedSection{{Label: "1.", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Uno"}}}}}},
		{Dictionary: dpdDictionary, Edition: dpdEdition, EntryID: "tilde2", Lemma: "tilde", CanonicalURL: dpdTildeURL, Sections: []parse.ParsedSection{{Label: "1.", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Dos"}}}}}},
	}}

	entries, _, err := NewDPDNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, result)
	if err != nil {
		t.Fatalf(dpdNormalizeErr, err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0].ID == entries[1].ID {
		t.Fatalf("entry ids must be distinct: %#v", entries)
	}
	if entries[0].ID != "dpd:tilde#tilde" || entries[1].ID != "dpd:tilde#tilde2" {
		t.Fatalf("entry ids = [%q %q]", entries[0].ID, entries[1].ID)
	}
}

func TestDPDNormalizerCoheresMixedBlocksIntoFallbackContent(t *testing.T) {
	result := parse.Result{Articles: []parse.ParsedArticle{{
		Dictionary:   dpdDictionary,
		Edition:      dpdEdition,
		EntryID:      "tilde",
		Lemma:        "tilde",
		CanonicalURL: dpdTildeURL,
		Sections: []parse.ParsedSection{{
			Label: "1.",
			Blocks: []parse.ParsedBlock{
				{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Antes"}},
				{Kind: parse.ParsedBlockKindTable, Table: &parse.ParsedTable{Headers: []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: dpdConTilde}, {HTML: "Sin tilde"}}}}, Rows: []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: "solo"}, {HTML: "solo"}}}}}},
				{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: dpdDespues}},
			},
		}},
	}}}

	entries, _, err := NewDPDNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, result)
	if err != nil {
		t.Fatalf(dpdNormalizeErr, err)
	}
	text := entries[0].Content
	for _, want := range []string{"1.", "Antes", "| Con tilde | Sin tilde |", "| solo      | solo      |", dpdDespues} {
		if !strings.Contains(text, want) {
			t.Fatalf("content missing %q\n%s", want, text)
		}
	}
}

func TestDPDNormalizerKeepsTableCellInlineMarkdownConsistent(t *testing.T) {
	table := normalizeTable(parse.ParsedTable{
		Headers: []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: `<em><span class="ment">Con</span></em> tilde`, Inlines: []model.Inline{{Kind: model.InlineKindEmphasis, Children: []model.Inline{{Kind: model.InlineKindMention, Text: "Con"}}}, {Kind: model.InlineKindText, Text: " tilde"}}}, {HTML: "Sin tilde"}}}},
		Rows:    []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: `<em><span class="ment">sólo</span></em>`, Inlines: []model.Inline{{Kind: model.InlineKindEmphasis, Children: []model.Inline{{Kind: model.InlineKindMention, Text: "sólo"}}}}}, {HTML: `solo <a href="tilde#n2">2</a>`, Inlines: []model.Inline{{Kind: model.InlineKindText, Text: "solo "}, {Kind: model.InlineKindReference, Text: "2", Target: "tilde#n2"}}}}}},
	})

	if got := table.Headers[0].Cells[0].Text; got != "*Con* tilde" {
		t.Fatalf("header cell = %q", got)
	}
	if got := table.Rows[0].Cells[0].Text; got != "*sólo*" {
		t.Fatalf("first row cell = %q", got)
	}
	if got := table.Rows[0].Cells[1].Text; got != "solo → [2](tilde#n2)" {
		t.Fatalf("second row cell = %q", got)
	}
	if len(table.Rows[0].Cells[1].Inlines) == 0 || table.Rows[0].Cells[1].Inlines[1].Kind != model.InlineKindReference {
		t.Fatalf("second row inline semantics = %#v", table.Rows[0].Cells[1].Inlines)
	}
}

func TestDPDNormalizerPreservesTableSpanMetadata(t *testing.T) {
	table := normalizeTable(parse.ParsedTable{
		Headers: []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: "Título", ColSpan: 4}}}},
		Rows:    []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: dpdConTilde, RowSpan: 2}, {HTML: "Caso"}, {HTML: "Ejemplo"}}}},
	})

	if got := table.Headers[0].Cells[0].ColSpan; got != 4 {
		t.Fatalf("header colspan = %d, want 4", got)
	}
	if got := table.Rows[0].Cells[0].RowSpan; got != 2 {
		t.Fatalf("row rowspan = %d, want 2", got)
	}
}

func TestDPDNormalizerUsesHTMLFallbackForComplexTablesInContentProjection(t *testing.T) {
	result := parse.Result{Articles: []parse.ParsedArticle{{
		Dictionary:   dpdDictionary,
		Edition:      dpdEdition,
		EntryID:      "tilde",
		Lemma:        "tilde",
		CanonicalURL: dpdTildeURL,
		Sections: []parse.ParsedSection{{
			Label: "3.2.1",
			Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindTable, Table: &parse.ParsedTable{
				Headers: []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: `<em>qué</em> / que`, ColSpan: 4, Inlines: []model.Inline{{Kind: model.InlineKindEmphasis, Text: "qué"}, {Kind: model.InlineKindText, Text: " / que"}}}}}},
				Rows:    []parse.ParsedTableRow{{Cells: []parse.ParsedTableCell{{HTML: dpdConTilde, RowSpan: 2}, {HTML: "Caso"}, {HTML: "Ejemplo"}}}},
			}}},
		}},
	}}}

	entries, _, err := NewDPDNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, result)
	if err != nil {
		t.Fatalf(dpdNormalizeErr, err)
	}
	text := entries[0].Content
	for _, want := range []string{"<table>", `<th colspan="4"><em>qué</em> / que</th>`, `<td rowspan="2">Con tilde</td>`} {
		if !strings.Contains(text, want) {
			t.Fatalf("content missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "|---") {
		t.Fatalf("content = %q, complex table must not degrade to markdown grid", text)
	}
}

func TestDPDNormalizerRespectsNestedScaffoldOverrideInMarkdown(t *testing.T) {
	paragraph := normalizeParagraph(parse.ParsedParagraph{Inlines: []model.Inline{{
		Kind: model.InlineKindEmphasis,
		Children: []model.Inline{{
			Kind: model.InlineKindMention,
			Children: []model.Inline{
				{Kind: model.InlineKindText, Text: "tilde"},
				{Kind: model.InlineKindScaffold, Text: "2"},
			},
		}},
	}}})

	if paragraph.Markdown != "*tilde* 2" {
		t.Fatalf("markdown = %q, want italic parent with plain override", paragraph.Markdown)
	}
}

func TestDPDNormalizerGluesNestedWordFragmentsAcrossInlineBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		input []model.Inline
		want  string
	}{
		{
			name: "mention plus scaffold keeps word intact",
			input: []model.Inline{{
				Kind: model.InlineKindMention,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "gr"},
					{Kind: model.InlineKindScaffold, Text: "úa"},
				},
			}},
			want: "*gr*úa",
		},
		{
			name: "nested plain and styled suffix stays glued",
			input: []model.Inline{{
				Kind: model.InlineKindMention,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "anch"},
					{Kind: model.InlineKindScaffold, Text: "oa"},
					{Kind: model.InlineKindCorrection, Text: "s"},
				},
			}},
			want: "*anch*oa*s*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paragraph := normalizeParagraph(parse.ParsedParagraph{Inlines: tt.input})
			if paragraph.Markdown != tt.want {
				t.Fatalf("markdown = %q, want %q", paragraph.Markdown, tt.want)
			}
		})
	}
}

func TestDPDNormalizerRejectsSyntheticQuoteNormalization(t *testing.T) {
	got := normalizeSingleParagraph(t, `Como adverbio significa <dfn>'correcta y adecuadamente'</dfn>.`)

	if got != "Como adverbio significa 'correcta y adecuadamente'." {
		t.Fatalf("normalizeParagraphMarkdown() = %q", got)
	}
}

func TestDPDNormalizerPreservesExampleAndEmphasisSemantics(t *testing.T) {
	got := normalizeSingleParagraph(t, `El comparativo es <em>mejor</em>. <span class="ejemplo">Cierra bien la ventana</span>.`)

	want := "El comparativo es *mejor*. ‹Cierra bien la ventana›."
	if got != want {
		t.Fatalf("normalizeParagraphMarkdown() = %q, want %q", got, want)
	}
}

func TestDPDNormalizerMarksRealBienExampleSemanticsExplicitly(t *testing.T) {
	got := normalizeSingleParagraph(t, `No he dormido bien esta noche: <span class="ejemplo">No he dormido bien esta noche</span>.`)

	if got != "No he dormido bien esta noche: ‹No he dormido bien esta noche›." {
		t.Fatalf("normalizeParagraphMarkdown() = %q", got)
	}
}

func TestDPDNormalizerPreservesStructuredInlineKinds(t *testing.T) {
	section := normalizeSections([]parse.ParsedSection{{
		Label: "1.",
		Blocks: []parse.ParsedBlock{{
			Kind: parse.ParsedBlockKindParagraph,
			Paragraph: &parse.ParsedParagraph{Inlines: []model.Inline{
				{Kind: model.InlineKindText, Text: "Debe evitarse "},
				{Kind: model.InlineKindMention, Text: "vide"},
				{Kind: model.InlineKindText, Text: ": "},
				{Kind: model.InlineKindCitationQuote, Children: []model.Inline{{Kind: model.InlineKindExclusion, Text: "⊗"}, {Kind: model.InlineKindText, Text: "Desde atrás vide..."}, {Kind: model.InlineKindBibliography, Text: "(González Dios mx 1999)"}}},
				{Kind: model.InlineKindText, Text: "."},
			}},
		}},
	}})

	paragraph := section[0].Paragraphs[0]
	if len(paragraph.Inlines) < 4 {
		t.Fatalf("inlines = %#v", paragraph.Inlines)
	}
	if paragraph.Inlines[1].Kind != model.InlineKindMention {
		t.Fatalf("mention kind = %q", paragraph.Inlines[1].Kind)
	}
	if paragraph.Inlines[3].Kind != model.InlineKindCitationQuote {
		t.Fatalf("citation kind = %q", paragraph.Inlines[3].Kind)
	}
	if paragraph.Markdown != "Debe evitarse *vide*: «⊗ Desde atrás vide... (González Dios mx 1999)»." {
		t.Fatalf("markdown = %q", paragraph.Markdown)
	}
}

func TestDPDNormalizerShapesCanonicalReferences(t *testing.T) {
	got := normalizeSingleParagraph(t, `Este uso no debe confundirse (→ <a href="bien#S1590507271213267522">6</a>).`)

	want := "Este uso no debe confundirse (→ [6](bien#S1590507271213267522))."
	if got != want {
		t.Fatalf("normalizeParagraphMarkdown() = %q, want %q", got, want)
	}
}

func TestDPDNormalizerKeepsIntegratedHeadingSemantics(t *testing.T) {
	normalized := normalizeSections([]parse.ParsedSection{{Label: "5.", Title: "bien que.", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: `Locución conjuntiva equivalente a <em>si bien</em>.`}}}}})

	if len(normalized) != 1 {
		t.Fatalf("sections = %d, want 1", len(normalized))
	}
	if normalized[0].Title != "bien que." {
		t.Fatalf("title = %q, want bien que.", normalized[0].Title)
	}
	if normalized[0].Paragraphs[0].Markdown != "Locución conjuntiva equivalente a *si bien*." {
		t.Fatalf("paragraph = %q", normalized[0].Paragraphs[0].Markdown)
	}
}

func TestDPDNormalizerSeparatesCitationFieldsFromProse(t *testing.T) {
	citation := normalizeCitation(parse.ParsedArticle{
		Dictionary: dpdDictionary,
		Edition:    dpdEdition,
		Citation:   parse.ParsedCitation{Text: "Real Academia Española y Asociación de Academias de la Lengua Española: Diccionario panhispánico de dudas (DPD) [en línea], https://www.rae.es/dpd/bien, 2.ª edición. [Consulta: 10/03/2026]."},
	})

	if citation.SourceLabel != "Real Academia Española y Asociación de Academias de la Lengua Española" {
		t.Fatalf("SourceLabel = %q", citation.SourceLabel)
	}
	if citation.CanonicalURL != "https://www.rae.es/dpd/bien" {
		t.Fatalf("CanonicalURL = %q", citation.CanonicalURL)
	}
	if citation.ConsultedAt != "10/03/2026" {
		t.Fatalf("ConsultedAt = %q", citation.ConsultedAt)
	}
}

func TestDPDSignsNormalizePhase1(t *testing.T) {
	if got := cleanInlineText("hello @ world"); got != "hello @ world" {
		t.Errorf("expected @ to survive, got %q", got)
	}
	if got := cleanInlineText("+ infinitivo"); got != "+ infinitivo" {
		t.Errorf("expected + to survive, got %q", got)
	}
	if got := normalizeSingleParagraph(t, `Tilde diacrítica en <em>qué</em> / que.`); got != "Tilde diacrítica en *qué* / que." {
		t.Errorf("expected slash to remain plain text, got %q", got)
	}
}

func TestDPDSignsNormalizePhase3Synthetic(t *testing.T) {
	// SYNTHETIC TEST - NO REAL HTML VALIDATION.
	// These cases are inferred and MUST be revisited when real DPD examples are found.
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "agrammatical marker survives", raw: "*", want: "*"},
		{name: "hypothetical marker survives", raw: "‖", want: "‖"},
		{name: "phoneme marker survives", raw: "//", want: "//"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanInlineText(tt.raw); got != tt.want {
				t.Fatalf("cleanInlineText(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
