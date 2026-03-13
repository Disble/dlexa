package normalize

import (
	"context"
	"strings"
	"testing"

	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/parse"
)

func normalizeSingleParagraph(t *testing.T, raw string) string {
	t.Helper()
	return normalizeParagraphMarkdown(raw)
}

func TestDPDNormalizerBuildsMarkdownReadyArticle(t *testing.T) {
	normalizer := NewDPDNormalizer()
	result := parse.Result{Articles: []parse.ParsedArticle{{
		Dictionary:   "Diccionario panhispánico de dudas",
		Edition:      "2.ª edición",
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
		t.Fatalf("Normalize() error = %v", err)
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
			{Kind: parse.ParsedBlockKindTable, Table: &parse.ParsedTable{Headers: []parse.ParsedTableRow{{Cells: []string{"A", "B"}}}, Rows: []parse.ParsedTableRow{{Cells: []string{"1", "2"}}}}},
			{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Después"}},
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
		{Dictionary: "Diccionario panhispánico de dudas", Edition: "2.ª edición", EntryID: "tilde", Lemma: "tilde", CanonicalURL: "https://www.rae.es/dpd/tilde", Sections: []parse.ParsedSection{{Label: "1.", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Uno"}}}}}},
		{Dictionary: "Diccionario panhispánico de dudas", Edition: "2.ª edición", EntryID: "tilde2", Lemma: "tilde", CanonicalURL: "https://www.rae.es/dpd/tilde", Sections: []parse.ParsedSection{{Label: "1.", Blocks: []parse.ParsedBlock{{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Dos"}}}}}},
	}}

	entries, _, err := NewDPDNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, result)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
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
		Dictionary:   "Diccionario panhispánico de dudas",
		Edition:      "2.ª edición",
		EntryID:      "tilde",
		Lemma:        "tilde",
		CanonicalURL: "https://www.rae.es/dpd/tilde",
		Sections: []parse.ParsedSection{{
			Label: "1.",
			Blocks: []parse.ParsedBlock{
				{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Antes"}},
				{Kind: parse.ParsedBlockKindTable, Table: &parse.ParsedTable{Headers: []parse.ParsedTableRow{{Cells: []string{"Con tilde", "Sin tilde"}}}, Rows: []parse.ParsedTableRow{{Cells: []string{"solo", "solo"}}}}},
				{Kind: parse.ParsedBlockKindParagraph, Paragraph: &parse.ParsedParagraph{HTML: "Después"}},
			},
		}},
	}}}

	entries, _, err := NewDPDNormalizer().Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, result)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	text := entries[0].Content
	for _, want := range []string{"1.", "Antes", "| Con tilde | Sin tilde |", "| solo      | solo      |", "Después"} {
		if !strings.Contains(text, want) {
			t.Fatalf("content missing %q\n%s", want, text)
		}
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
		Dictionary: "Diccionario panhispánico de dudas",
		Edition:    "2.ª edición",
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
