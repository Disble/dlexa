package render

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

const (
	testDictionary     = "Diccionario panhispánico de dudas"
	testEdition        = "2.ª edición"
	testSourceLabel    = "Real Academia Española y Asociación de Academias de la Lengua Española"
	testConsultedAt    = "10/03/2026"
	testColSinTilde    = "Sin tilde"
	testColConTilde    = "Con tilde"
	testErrRender      = "Render() error = %v"
	testPayloadMissing = "payload missing %q\n%s"
	testTildeURL       = "https://www.rae.es/dpd/tilde"
	testSoloToTildeURL = "https://www.rae.es/dpd/solo → https://www.rae.es/dpd/tilde"
)

var reANSITest = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func paragraphBlock(markdown string, inlines ...model.Inline) model.Block {
	paragraph := model.Paragraph{Markdown: markdown, Inlines: inlines}
	return model.Block{Kind: model.ArticleBlockKindParagraph, Paragraph: &paragraph}
}

func tableBlock(headers, rows [][]string) model.Block {
	makeRows := func(raw [][]string) []model.TableRow {
		out := make([]model.TableRow, 0, len(raw))
		for _, row := range raw {
			cells := make([]model.TableCell, 0, len(row))
			for _, cell := range row {
				cells = append(cells, model.TableCell{Text: cell})
			}
			out = append(out, model.TableRow{Cells: cells})
		}
		return out
	}
	table := model.Table{Headers: makeRows(headers), Rows: makeRows(rows)}
	return model.Block{Kind: model.ArticleBlockKindTable, Table: &table}
}

func inlineTableBlock(headers, rows [][]model.TableCell) model.Block {
	makeRows := func(raw [][]model.TableCell) []model.TableRow {
		out := make([]model.TableRow, 0, len(raw))
		for _, row := range raw {
			cells := make([]model.TableCell, 0, len(row))
			cells = append(cells, row...)
			out = append(out, model.TableRow{Cells: cells})
		}
		return out
	}
	table := model.Table{Headers: makeRows(headers), Rows: makeRows(rows)}
	return model.Block{Kind: model.ArticleBlockKindTable, Table: &table}
}

func spanningTableBlock(headers, rows [][]model.TableCell) model.Block {
	return inlineTableBlock(headers, rows)
}

func sampleBienArticle() *model.Article {
	sections := []model.Section{
		{Label: "1.", Blocks: []model.Block{paragraphBlock("Como adverbio de modo significa 'correcta y adecuadamente': *Cierra bien la ventana, por favor*; 'satisfactoriamente': *No he dormido bien esta noche*. El comparativo es *mejor*. No debe usarse *más bien* como comparativo. Este uso incorrecto no debe confundirse con los usos correctos de la locución adverbial *más bien* (→ [6](bien#S1590507271213267522)).", model.Inline{Kind: model.InlineKindText, Text: "Como adverbio de modo significa "}, model.Inline{Kind: model.InlineKindGloss, Text: "'correcta y adecuadamente'"}, model.Inline{Kind: model.InlineKindText, Text: ": "}, model.Inline{Kind: model.InlineKindExample, Text: "Cierra bien la ventana, por favor"}, model.Inline{Kind: model.InlineKindText, Text: "; 'satisfactoriamente': "}, model.Inline{Kind: model.InlineKindExample, Text: "No he dormido bien esta noche"}, model.Inline{Kind: model.InlineKindText, Text: ". El comparativo es "}, model.Inline{Kind: model.InlineKindMention, Text: "mejor"}, model.Inline{Kind: model.InlineKindText, Text: ". No debe usarse "}, model.Inline{Kind: model.InlineKindMention, Text: "más bien"}, model.Inline{Kind: model.InlineKindText, Text: " como comparativo. Este uso incorrecto no debe confundirse con los usos correctos de la locución adverbial "}, model.Inline{Kind: model.InlineKindMention, Text: "más bien"}, model.Inline{Kind: model.InlineKindText, Text: " ("}, model.Inline{Kind: model.InlineKindReference, Text: "6", Target: "bien#S1590507271213267522"}, model.Inline{Kind: model.InlineKindText, Text: ")."})}},
		{Label: "2.", Blocks: []model.Block{paragraphBlock("Antepuesto a un adjetivo o a otro adverbio, funciona como intensificador enfático, con valor equivalente a *muy*.", model.Inline{Kind: model.InlineKindText, Text: "Antepuesto a un adjetivo o a otro adverbio, funciona como intensificador enfático, con valor equivalente a "}, model.Inline{Kind: model.InlineKindEmphasis, Text: "muy"}, model.Inline{Kind: model.InlineKindText, Text: "."})}},
		{Label: "3.", Blocks: []model.Block{paragraphBlock("Repetido ante dos o más elementos de una oración, señala distintas posibilidades.")}},
		{Label: "4.", Blocks: []model.Block{paragraphBlock("Como adjetivo invariable significa 'de buena posición social'.")}},
		{Label: "5.", Title: "bien que.", Blocks: []model.Block{paragraphBlock("Locución conjuntiva concesiva equivalente a 'aunque'. Con este mismo sentido, se emplea más frecuentemente la locución *si bien* (→ [7](bien#S1590507271244936818)).", model.Inline{Kind: model.InlineKindText, Text: "Locución conjuntiva concesiva equivalente a 'aunque'. Con este mismo sentido, se emplea más frecuentemente la locución "}, model.Inline{Kind: model.InlineKindMention, Text: "si bien"}, model.Inline{Kind: model.InlineKindText, Text: " ("}, model.Inline{Kind: model.InlineKindReference, Text: "7", Target: "bien#S1590507271244936818"}, model.Inline{Kind: model.InlineKindText, Text: ")."})}},
		{Label: "6.", Title: "más bien.", Blocks: []model.Block{paragraphBlock("Locución adverbial que se usa con distintos valores:")}, Children: []model.Section{{Label: "a)", Blocks: []model.Block{paragraphBlock("Para introducir una rectificación o una matización.")}}, {Label: "b)", Blocks: []model.Block{paragraphBlock("Con el sentido de 'en cierto modo, de alguna manera'.")}}, {Label: "c)", Blocks: []model.Block{paragraphBlock("También significa 'mejor o preferentemente'.")}}}},
		{Label: "7.", Title: "si bien.", Blocks: []model.Block{paragraphBlock("Locución conjuntiva concesiva equivalente a 'aunque'. Con este mismo sentido, se emplea también, aunque con menos frecuencia, la locución *bien que* (→ [5](bien#S1590507271206059324)).", model.Inline{Kind: model.InlineKindText, Text: "Locución conjuntiva concesiva equivalente a 'aunque'. Con este mismo sentido, se emplea también, aunque con menos frecuencia, la locución "}, model.Inline{Kind: model.InlineKindMention, Text: "bien que"}, model.Inline{Kind: model.InlineKindText, Text: " ("}, model.Inline{Kind: model.InlineKindReference, Text: "5", Target: "bien#S1590507271206059324"}, model.Inline{Kind: model.InlineKindText, Text: ")."})}},
	}
	for i := range sections {
		sections[i].Paragraphs = paragraphsFromBlocks(sections[i].Blocks)
	}
	return &model.Article{
		Dictionary: testDictionary,
		Edition:    testEdition,
		Sections:   sections,
		Citation: model.Citation{
			SourceLabel:  testSourceLabel,
			CanonicalURL: "https://www.rae.es/dpd/bien",
			Edition:      testEdition,
			ConsultedAt:  testConsultedAt,
		},
	}
}

func sampleTildeEntries() []model.Entry {
	first := model.Article{
		Dictionary: testDictionary,
		Edition:    testEdition,
		Lemma:      "tilde",
		Sections: []model.Section{{
			Label: "1.",
			Blocks: []model.Block{
				paragraphBlock("La tilde permite distinguir ciertos usos en la escritura actual."),
				tableBlock([][]string{{testColConTilde, testColSinTilde}}, [][]string{{"aún", "aun"}, {"sólo", "solo"}}),
				paragraphBlock("La recomendación vigente depende del contexto y de la ambigüedad."),
			},
		}},
		Citation: model.Citation{SourceLabel: testSourceLabel, CanonicalURL: testTildeURL, Edition: testEdition, ConsultedAt: testConsultedAt},
	}
	second := model.Article{
		Dictionary: testDictionary,
		Edition:    testEdition,
		Lemma:      "tilde",
		Sections: []model.Section{{
			Label:  "2.",
			Blocks: []model.Block{paragraphBlock("En casos enfáticos o diacríticos, la grafía histórica puede mantenerse como referencia."), paragraphBlock("La norma vigente prioriza la claridad sin multiplicar tildes innecesarias.")},
		}},
		Citation: model.Citation{SourceLabel: testSourceLabel, CanonicalURL: testTildeURL, Edition: testEdition, ConsultedAt: testConsultedAt},
	}
	for i := range first.Sections {
		first.Sections[i].Paragraphs = paragraphsFromBlocks(first.Sections[i].Blocks)
	}
	for i := range second.Sections {
		second.Sections[i].Paragraphs = paragraphsFromBlocks(second.Sections[i].Blocks)
	}
	return []model.Entry{{ID: "dpd:tilde#tilde", Headword: "tilde", Article: &first}, {ID: "dpd:tilde#tilde2", Headword: "tilde", Article: &second}}
}

func paragraphsFromBlocks(blocks []model.Block) []model.Paragraph {
	paragraphs := make([]model.Paragraph, 0, len(blocks))
	for _, block := range blocks {
		if block.Kind == model.ArticleBlockKindParagraph && block.Paragraph != nil {
			paragraphs = append(paragraphs, *block.Paragraph)
		}
	}
	return paragraphs
}

func TestMarkdownRendererRendersDPDArticleGolden(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "bien", Format: "markdown"},
		Entries: []model.Entry{{Headword: "bien", Article: sampleBienArticle()}},
	})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}

	want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", "bien.md.golden"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if stripANSITestOutput(string(payload)) != string(want) {
		t.Fatalf("Render() mismatch\n--- got ---\n%s\n--- want ---\n%s", payload, want)
	}
}

func TestMarkdownRendererRendersGroupedTildeArticlesAndTables(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "tilde", Format: "markdown"},
		Entries: sampleTildeEntries(),
	})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}

	want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", "tilde.md.golden"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if stripANSITestOutput(string(payload)) != string(want) {
		t.Fatalf("Render() mismatch\n--- got ---\n%s\n--- want ---\n%s", payload, want)
	}
}

func TestMarkdownRendererKeepsTableCellsMarkdownOnly(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "tilde",
		Article: &model.Article{Sections: []model.Section{{
			Label: "1.",
			Blocks: []model.Block{inlineTableBlock(
				[][]model.TableCell{{
					{Text: "*Con* tilde", Inlines: []model.Inline{{Kind: model.InlineKindMention, Text: "Con"}, {Kind: model.InlineKindText, Text: " tilde"}}},
					{Text: testColSinTilde},
				}},
				[][]model.TableCell{{
					{Text: "*sólo*", Inlines: []model.Inline{{Kind: model.InlineKindMention, Text: "sólo"}}},
					{Text: "solo → [2](tilde#n2)", Inlines: []model.Inline{{Kind: model.InlineKindText, Text: "solo "}, {Kind: model.InlineKindReference, Text: "2", Target: "tilde#n2"}}},
				}},
			)},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	for _, want := range []string{"| *Con* tilde |", "| *sólo*      | solo → [2](tilde#n2)"} {
		if !strings.Contains(text, want) {
			t.Fatalf(testPayloadMissing, want, text)
		}
	}
	for _, forbidden := range []string{"<em>", "<span", "<a href"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("payload leaked html %q\n%s", forbidden, text)
		}
	}
}

func TestMarkdownRendererUsesValidMarkdownDividerForSimpleTables(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "tilde",
		Article: &model.Article{Sections: []model.Section{{
			Label:  "1.",
			Blocks: []model.Block{tableBlock([][]string{{testColConTilde, testColSinTilde}}, [][]string{{"aún", "aun"}})},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	if !strings.Contains(text, "|-----------|-----------|") {
		t.Fatalf("payload = %q, want pipe-only markdown divider", payload)
	}
	if strings.Contains(text, "+") {
		t.Fatalf("payload = %q, markdown table divider must not use '+'", payload)
	}
}

func TestMarkdownRendererFallsBackToHTMLForComplexTables(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "tilde",
		Article: &model.Article{Sections: []model.Section{{
			Label: "3.2.1",
			Blocks: []model.Block{spanningTableBlock(
				[][]model.TableCell{{
					{Text: "Tilde diacrítica en *qué* / que", ColSpan: 4},
				}},
				[][]model.TableCell{
					{{Text: testColConTilde, RowSpan: 2}, {Text: "Con valor interrogativo o exclamativo", RowSpan: 2}, {Text: "Encabezando estructuras"}, {Text: "*¿Qué* calor!"}},
					{{Text: "Sustantivados"}, {Text: "el *cuándo*"}},
				},
			)},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	for _, want := range []string{"<table>", "<th colspan=\"4\">Tilde diacrítica en <em>qué</em> / que</th>", "<td rowspan=\"2\">Con tilde</td>", "<td><em>¿Qué</em> calor!</td>"} {
		if !strings.Contains(text, want) {
			t.Fatalf(testPayloadMissing, want, text)
		}
	}
	if strings.Contains(text, "|-----------|") {
		t.Fatalf("payload = %q, complex tables must not render as markdown grid", payload)
	}
}

func TestMarkdownRendererRespectsNestedFormattingOverride(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "tilde",
		Article: &model.Article{Sections: []model.Section{{
			Label: "1.",
			Blocks: []model.Block{paragraphBlock("", model.Inline{
				Kind: model.InlineKindEmphasis,
				Children: []model.Inline{{
					Kind: model.InlineKindMention,
					Children: []model.Inline{
						{Kind: model.InlineKindText, Text: "tilde"},
						{Kind: model.InlineKindScaffold, Text: "2"},
					},
				}},
			})},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	if !strings.Contains(text, "1. *tilde* 2") {
		t.Fatalf("payload = %q, want italic root with plain nested override", payload)
	}
	if strings.Contains(text, "*tilde 2*") {
		t.Fatalf("payload = %q, nested override leaked parent italic", payload)
	}
}

func TestMarkdownRendererGluesNestedWordFragmentsAcrossInlineBoundaries(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "grua",
		Article: &model.Article{Sections: []model.Section{{
			Label: "1.",
			Blocks: []model.Block{paragraphBlock("",
				model.Inline{Kind: model.InlineKindMention, Children: []model.Inline{{Kind: model.InlineKindText, Text: "gr"}, {Kind: model.InlineKindScaffold, Text: "úa"}}},
				model.Inline{Kind: model.InlineKindText, Text: ". "},
				model.Inline{Kind: model.InlineKindMention, Children: []model.Inline{{Kind: model.InlineKindText, Text: "anch"}, {Kind: model.InlineKindScaffold, Text: "oa"}, {Kind: model.InlineKindCorrection, Text: "s"}}},
			)},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	if !strings.Contains(text, "1. *gr*úa. *anch*oa*s*") {
		t.Fatalf("payload = %q, want glued lexical fragments across formatting boundaries", payload)
	}
	for _, bad := range []string{"*gr* úa", "*anch* oa *s*"} {
		if strings.Contains(text, bad) {
			t.Fatalf("payload = %q, contains split lexical fragment %q", payload, bad)
		}
	}
}

func TestMarkdownRendererUsesPlainMarkdownGroupedHeading(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{
		Request: model.LookupRequest{Query: "tilde", Format: "markdown"},
		Entries: []model.Entry{{
			ID:       "dpd:tilde#tilde2",
			Headword: "tilde2",
			Article: &model.Article{Lemma: "tilde2", Sections: []model.Section{{
				Label:  "1.",
				Blocks: []model.Block{paragraphBlock("Variante")},
			}}},
		}},
	})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	if !strings.Contains(text, "# tilde2") {
		t.Fatalf("payload = %q, want markdown-only heading", payload)
	}
	for _, forbidden := range []string{"<span", "<sup>"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("payload leaked html heading %q\n%s", forbidden, text)
		}
	}
}

func TestMarkdownRendererRendersPreservedQuotes(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "bien",
		Article: &model.Article{Sections: []model.Section{{
			Label:      "1.",
			Blocks:     []model.Block{paragraphBlock("Como adverbio significa 'correcta y adecuadamente'.")},
			Paragraphs: []model.Paragraph{{Markdown: "Como adverbio significa 'correcta y adecuadamente'."}},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	if !strings.Contains(string(payload), "'correcta y adecuadamente'") {
		t.Fatalf("payload = %q, want authored quotes", payload)
	}
}

func TestMarkdownRendererRendersSemanticMarkdownOutput(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "bien",
		Article: &model.Article{Sections: []model.Section{{
			Label: "1.",
			Blocks: []model.Block{paragraphBlock(
				"El comparativo es *mejor*. *Cierra bien la ventana*.",
				model.Inline{Kind: model.InlineKindText, Text: "El comparativo es "},
				model.Inline{Kind: model.InlineKindMention, Text: "mejor"},
				model.Inline{Kind: model.InlineKindText, Text: ". "},
				model.Inline{Kind: model.InlineKindExample, Text: "Cierra bien la ventana"},
				model.Inline{Kind: model.InlineKindText, Text: "."},
			)},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	if strings.Contains(text, "\x1b[") {
		t.Fatalf("payload = %q, markdown output must stay ANSI-free", payload)
	}
	for _, want := range []string{"El comparativo es *mejor*.", "*Cierra bien la ventana*."} {
		if !strings.Contains(text, want) {
			t.Fatalf("payload = %q, missing %q", payload, want)
		}
	}
	for _, forbidden := range []string{"[ej.:", "ej.:", "‹", "›"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("payload = %q, contains forbidden wrapper %q", payload, forbidden)
		}
	}
}

func TestMarkdownRendererKeepsRealBienExampleRecoverableInMarkdownOutput(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{Headword: "bien", Article: sampleBienArticle()}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}

	text := string(payload)
	for _, want := range []string{"*No he dormido bien esta noche*", "*Cierra bien la ventana, por favor*"} {
		if !strings.Contains(text, want) {
			t.Fatalf("payload missing emphasized bien example %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "'satisfactoriamente': No he dormido bien esta noche.") {
		t.Fatalf("payload collapsed example into ordinary prose\n%s", text)
	}
}

func TestMarkdownRendererUsesMarkdownSyntaxWithoutInventedFallbacks(t *testing.T) {
	article := &model.Article{Sections: []model.Section{{Label: "1.", Blocks: []model.Block{paragraphBlock("", model.Inline{Kind: model.InlineKindText, Text: "El comparativo es "}, model.Inline{Kind: model.InlineKindMention, Text: "mejor"}, model.Inline{Kind: model.InlineKindText, Text: ": "}, model.Inline{Kind: model.InlineKindExample, Text: "Cierra bien la ventana"})}}}}

	for name, renderer := range map[string]*MarkdownRenderer{
		"default": NewMarkdownRenderer(),
		"plain":   NewMarkdownRendererWithProfile(TerminalProfile{ANSIEnabled: false}),
		"rich":    NewMarkdownRendererWithProfile(TerminalProfile{ANSIEnabled: true}),
	} {
		payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{Headword: "bien", Article: article}}})
		if err != nil {
			t.Fatalf("%s Render() error = %v", name, err)
		}
		text := string(payload)
		if strings.Contains(text, "\x1b[") {
			t.Fatalf("%s payload contains ANSI bytes\n%s", name, text)
		}
		for _, want := range []string{"El comparativo es *mejor*:", "*Cierra bien la ventana*"} {
			if !strings.Contains(text, want) {
				t.Fatalf("%s payload = %q, missing %q", name, payload, want)
			}
		}
		for _, forbidden := range []string{"[ej.:", "ej.:", "‹", "›"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s payload = %q, contains forbidden marker %q", name, payload, forbidden)
			}
		}
	}
}

func TestMarkdownRendererRendersSingleArrowCrossReferences(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "bien",
		Article: &model.Article{Sections: []model.Section{{
			Label:  "1.",
			Blocks: []model.Block{paragraphBlock("Ver (→ [6](bien#S1590507271213267522)) y → [7](bien#S1590507271244936818).")},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	if strings.Contains(text, "[→ 6]") || strings.Contains(text, "(→ [→ 6]") {
		t.Fatalf("payload = %q, got malformed duplicated reference", payload)
	}
	if strings.Count(text, "→ [6](bien#S1590507271213267522)") != 1 || strings.Count(text, "→ [7](bien#S1590507271244936818)") != 1 {
		t.Fatalf("payload = %q, want exactly one markdown arrow reference per section", payload)
	}
}

func TestMarkdownRendererRendersIntegratedLexicalHeads(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "bien",
		Article: &model.Article{Sections: []model.Section{{
			Label:  "5.",
			Title:  "bien que.",
			Blocks: []model.Block{paragraphBlock("Locución conjuntiva concesiva.")},
		}}},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	if !strings.Contains(text, "5. bien que. Locución conjuntiva concesiva.") {
		t.Fatalf("payload = %q, want integrated lexical head", payload)
	}
}

func TestMarkdownRendererKeepsSectionAndSubitemLayoutCoherent(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{Headword: "bien", Article: sampleBienArticle()}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}

	text := string(payload)
	for _, bad := range []string{"\n\n5. bien que.\n\n", "\n\n6. más bien.\n\n", "\n\na)\n\n", "\n\nb)\n\n", "\n\nc)\n\n"} {
		if strings.Contains(text, bad) {
			t.Fatalf("payload contains fragmented layout %q\n%s", bad, text)
		}
	}
	for _, want := range []string{"6. más bien. Locución adverbial que se usa con distintos valores:", "  a) Para introducir una rectificación o una matización.", "  b) Con el sentido de 'en cierto modo, de alguna manera'.", "  c) También significa 'mejor o preferentemente'."} {
		if !strings.Contains(text, want) {
			t.Fatalf("payload missing coherent layout %q\n%s", want, text)
		}
	}
}

func TestMarkdownRendererRendersStructuredCitation(t *testing.T) {
	renderer := NewMarkdownRenderer()
	payload, err := renderer.Render(context.Background(), model.LookupResult{Entries: []model.Entry{{
		Headword: "bien",
		Article: &model.Article{
			Dictionary: testDictionary,
			Citation: model.Citation{
				SourceLabel:  testSourceLabel,
				Edition:      testEdition,
				CanonicalURL: "https://www.rae.es/dpd/bien",
				ConsultedAt:  testConsultedAt,
			},
		},
	}}})
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	for _, want := range []string{"Source: Real Academia Española y Asociación de Academias de la Lengua Española", "Dictionary: Diccionario panhispánico de dudas", "Edition: 2.ª edición", "URL: https://www.rae.es/dpd/bien", "Consulted: 10/03/2026"} {
		if !strings.Contains(text, want) {
			t.Fatalf(testPayloadMissing, want, text)
		}
	}
}

func TestMarkdownRendererRendersStructuredLookupMissGuidance(t *testing.T) {
	renderer := NewMarkdownRenderer()
	tests := []struct {
		name     string
		result   model.LookupResult
		mustHave []string
		mustNot  []string
	}{
		{
			name: "native suggestion surfaces without search nudge",
			result: model.LookupResult{
				Request: model.LookupRequest{Query: "alicuota", Format: "markdown"},
				Misses: []model.LookupMiss{{
					Kind:       model.LookupMissKindRelatedEntry,
					Query:      "alicuota",
					Source:     "dpd",
					Suggestion: &model.LookupSuggestion{Kind: "related_entry", DisplayText: "alícuota", URL: "https://www.rae.es/dpd/alícuota"},
				}},
			},
			mustHave: []string{"# alicuota", "Quizá quiso decir **alícuota**.", "https://www.rae.es/dpd/alícuota"},
			mustNot:  []string{"dlexa search alicuota"},
		},
		{
			name: "generic miss nudges explicit search command only",
			result: model.LookupResult{
				Request: model.LookupRequest{Query: "zumbidoinexistente", Format: "markdown"},
				Misses: []model.LookupMiss{{
					Kind:       model.LookupMissKindGenericNotFound,
					Query:      "zumbidoinexistente",
					Source:     "dpd",
					NextAction: &model.LookupNextAction{Kind: model.LookupNextActionKindSearch, Query: "zumbidoinexistente", Command: "dlexa search zumbidoinexistente"},
				}},
			},
			mustHave: []string{"# zumbidoinexistente", "Try `dlexa search zumbidoinexistente`."},
			mustNot:  []string{"Quizá quiso decir", "automatic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := renderer.Render(context.Background(), tt.result)
			if err != nil {
				t.Fatalf(testErrRender, err)
			}
			assertMarkdownPayloadGuidance(t, string(payload), tt.mustHave, tt.mustNot)
		})
	}
}

func assertMarkdownPayloadGuidance(t *testing.T, text string, mustHave []string, mustNot []string) {
	t.Helper()
	for _, want := range mustHave {
		if !strings.Contains(text, want) {
			t.Fatalf(testPayloadMissing, want, text)
		}
	}
	for _, forbidden := range mustNot {
		if strings.Contains(text, forbidden) {
			t.Fatalf("payload contains forbidden text %q\n%s", forbidden, text)
		}
	}
}

func TestMarkdownRendererShowsRedirectWarningAboveArticle(t *testing.T) {
	result := model.LookupResult{
		Request:  model.LookupRequest{Query: "solo"},
		Warnings: []model.Warning{{Code: model.WarningCodeDPDRedirected, Message: testSoloToTildeURL, Source: "dpd"}},
		Entries: []model.Entry{
			{
				ID:       "tilde1",
				Headword: "tilde1",
				Article: &model.Article{
					Dictionary: testDictionary,
					Edition:    testEdition,
					Lemma:      "tilde1",
					Sections: []model.Section{
						{Blocks: []model.Block{paragraphBlock("Definición de tilde.")}},
					},
				},
			},
		},
	}

	payload, err := NewMarkdownRenderer().Render(context.Background(), result)
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)

	// Warning must appear before the article content
	warnIdx := strings.Index(text, "⚠")
	articleIdx := strings.Index(text, "tilde1")
	if warnIdx == -1 {
		t.Fatal("redirect warning block missing from output")
	}
	if articleIdx == -1 {
		t.Fatal("article content missing from output")
	}
	if warnIdx > articleIdx {
		t.Fatalf("redirect warning appears after article content (warn at %d, article at %d)\n%s", warnIdx, articleIdx, text)
	}

	mustHave := []string{
		testSoloToTildeURL,
		"\"solo\"",
		"tilde1",
	}
	mustNot := []string{"solo\n"}
	assertMarkdownPayloadGuidance(t, text, mustHave, mustNot)
}

func TestMarkdownRendererDoesNotShowRedirectBlockWhenNoRedirect(t *testing.T) {
	result := model.LookupResult{
		Request: model.LookupRequest{Query: "bien"},
		Entries: []model.Entry{
			{
				ID:       "bien",
				Headword: "bien",
				Article: &model.Article{
					Dictionary: testDictionary,
					Edition:    testEdition,
					Lemma:      "bien",
					Sections: []model.Section{
						{Blocks: []model.Block{paragraphBlock("Definición de bien.")}},
					},
				},
			},
		},
	}

	payload, err := NewMarkdownRenderer().Render(context.Background(), result)
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	assertMarkdownPayloadGuidance(t, string(payload), []string{"bien"}, []string{"⚠", "redirige"})
}

func stripANSITestOutput(text string) string {
	return reANSITest.ReplaceAllString(text, "")
}

func TestMarkdownRendererShowsRedirectURLInCitation(t *testing.T) {
	result := model.LookupResult{
		Request:  model.LookupRequest{Query: "solo"},
		Warnings: []model.Warning{{Code: model.WarningCodeDPDRedirected, Message: testSoloToTildeURL, Source: "dpd"}},
		Entries: []model.Entry{
			{
				ID:       "tilde1",
				Headword: "tilde1",
				Article: &model.Article{
					Dictionary: testDictionary,
					Citation: model.Citation{
						SourceLabel:  testSourceLabel,
						Edition:      testEdition,
						CanonicalURL: testTildeURL,
						ConsultedAt:  testConsultedAt,
					},
				},
			},
		},
	}

	payload, err := NewMarkdownRenderer().Render(context.Background(), result)
	if err != nil {
		t.Fatalf(testErrRender, err)
	}
	text := string(payload)
	assertMarkdownPayloadGuidance(t, text,
		[]string{"URL: " + testSoloToTildeURL},
		[]string{"URL: " + testTildeURL + "\n"}, // must NOT appear alone
	)
}
