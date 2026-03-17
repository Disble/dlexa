package parse

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

func loadDPDFixtureHTML(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "dpd", name+".html")) //nolint:gosec // G304: test fixture, path not from user input
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return body
}

func parseFixture(t *testing.T, name string) []ParsedArticle {
	t.Helper()
	parser := NewDPDArticleParser()
	result, _, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd", DisplayName: name}, fetch.Document{
		URL:         "https://www.rae.es/dpd/" + name,
		ContentType: "text/html; charset=utf-8",
		StatusCode:  200,
		Body:        loadDPDFixtureHTML(t, name),
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	return result.Articles
}

func assertBienArticleMetadata(t *testing.T, article ParsedArticle) {
	t.Helper()
	if article.EntryID != "bien" {
		t.Fatalf("EntryID = %q", article.EntryID)
	}
	if article.Dictionary != "Diccionario panhispánico de dudas" {
		t.Fatalf("Dictionary = %q", article.Dictionary)
	}
	if article.Edition != "2.ª edición" {
		t.Fatalf("Edition = %q", article.Edition)
	}
	if article.Lemma != "bien" {
		t.Fatalf("Lemma = %q", article.Lemma)
	}
}

func assertBienArticleSections(t *testing.T, article ParsedArticle) {
	t.Helper()
	if len(article.Sections) != 7 {
		t.Fatalf("top-level sections = %d, want 7", len(article.Sections))
	}
	if len(article.Sections[5].Children) != 3 {
		t.Fatalf("section 6 children = %d, want 3", len(article.Sections[5].Children))
	}
	if article.Sections[4].Title != "bien que." || article.Sections[5].Title != "más bien." || article.Sections[6].Title != "si bien." {
		t.Fatalf("lexical titles = [%q %q %q], want bien que./más bien./si bien.", article.Sections[4].Title, article.Sections[5].Title, article.Sections[6].Title)
	}
}

func assertBienFirstParagraph(t *testing.T, article ParsedArticle) {
	t.Helper()
	joined := article.Sections[0].Paragraphs[0].HTML
	if strings.Contains(joined, "La institución") || strings.Contains(joined, "Boletín de novedades") || strings.Contains(joined, "Qué contiene") {
		t.Fatalf("article paragraph leaked chrome: %q", joined)
	}
	if !strings.Contains(joined, "<em><span class=\"ment\">mejor</span></em>") {
		t.Fatalf("article paragraph missing emphasis span: %q", joined)
	}
	if !strings.Contains(joined, "(→ <a href=\"bien#S1590507271213267522\"") {
		t.Fatalf("article paragraph missing source reference semantics: %q", joined)
	}
}

func TestDPDArticleParserExtractsBienArticleAndSkipsChrome(t *testing.T) {
	parser := NewDPDArticleParser()
	result, warnings, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd", DisplayName: "bien"}, fetch.Document{
		URL:         "https://www.rae.es/dpd/bien",
		ContentType: "text/html; charset=utf-8",
		StatusCode:  200,
		Body:        loadDPDFixtureHTML(t, "bien"),
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(warnings) == 0 || warnings[0].Code != "dpd_access_profile" {
		t.Fatalf("Parse() warnings = %#v, want access profile warning", warnings)
	}
	if len(result.Articles) != 1 {
		t.Fatalf("Parse() articles = %d, want 1", len(result.Articles))
	}

	article := result.Articles[0]
	assertBienArticleMetadata(t, article)
	assertBienArticleSections(t, article)
	assertBienFirstParagraph(t, article)
	if !strings.Contains(article.Citation.Text, "Diccionario panhispánico de dudas") {
		t.Fatalf("citation text = %q", article.Citation.Text)
	}
}

func assertTildeFirstSection(t *testing.T, first ParsedSection) {
	t.Helper()
	if len(first.Blocks) != 3 {
		t.Fatalf("first section blocks = %d, want 3", len(first.Blocks))
	}
	if first.Blocks[0].Kind != ParsedBlockKindParagraph || first.Blocks[1].Kind != ParsedBlockKindTable || first.Blocks[2].Kind != ParsedBlockKindParagraph {
		t.Fatalf("first section block kinds = %#v", first.Blocks)
	}
	if len(first.Paragraphs) != 2 {
		t.Fatalf("paragraph projection = %d, want 2", len(first.Paragraphs))
	}
}

func assertTildeTable(t *testing.T, table *ParsedTable) {
	t.Helper()
	if table == nil {
		t.Fatal("table block = nil")
	}
	if len(table.Headers) != 1 || len(table.Headers[0].Cells) != 2 {
		t.Fatalf("headers = %#v", table.Headers)
	}
	if table.Headers[0].Cells[0].HTML != "Con tilde" || table.Headers[0].Cells[1].HTML != "Sin tilde" {
		t.Fatalf("header cells = %#v", table.Headers[0].Cells)
	}
	if len(table.Rows) != 2 || table.Rows[1].Cells[1].HTML != "solo" {
		t.Fatalf("rows = %#v", table.Rows)
	}
	if table.Headers[0].Cells[0].ColSpan != 0 || table.Rows[0].Cells[0].RowSpan != 0 {
		t.Fatalf("simple table should not invent spans: %#v %#v", table.Headers[0].Cells[0], table.Rows[0].Cells[0])
	}
}

func TestDPDArticleParserExtractsMultiEntryTildeWithMixedBlocks(t *testing.T) {
	articles := parseFixture(t, "tilde")
	if len(articles) != 2 {
		t.Fatalf("articles = %d, want 2", len(articles))
	}
	if articles[0].Lemma != "tilde" || articles[1].Lemma != "tilde" {
		t.Fatalf("lemmas = [%q %q], want duplicated tilde entries", articles[0].Lemma, articles[1].Lemma)
	}
	if articles[0].EntryID != "tilde" || articles[1].EntryID != "tilde2" {
		t.Fatalf("entry ids = [%q %q]", articles[0].EntryID, articles[1].EntryID)
	}

	first := articles[0].Sections[0]
	assertTildeFirstSection(t, first)
	assertTildeTable(t, first.Blocks[1].Table)

	second := articles[1].Sections[0]
	if len(second.Blocks) != 2 || second.Blocks[0].Kind != ParsedBlockKindParagraph || second.Blocks[1].Kind != ParsedBlockKindParagraph {
		t.Fatalf("second entry blocks = %#v", second.Blocks)
	}
	if !strings.Contains(second.Blocks[1].Paragraph.HTML, "vigente") {
		t.Fatalf("second entry paragraph = %#v", second.Blocks[1].Paragraph)
	}
}

func TestParseTablePreservesRowAndColumnSpans(t *testing.T) {
	table := parseTable(`<table><tr><th colspan="4">Tilde diacrítica en <em>qué</em> / que</th></tr><tr><td rowspan="2">Con tilde</td><td>Caso</td><td>Ejemplo</td></tr><tr><td colspan="2"><em>¿Qué</em> calor!</td></tr></table>`)
	if table == nil {
		t.Fatal("table = nil")
	}
	if got := table.Headers[0].Cells[0].ColSpan; got != 4 {
		t.Fatalf("header colspan = %d, want 4", got)
	}
	if got := table.Rows[0].Cells[0].RowSpan; got != 2 {
		t.Fatalf("rowspan = %d, want 2", got)
	}
	if got := table.Rows[1].Cells[0].ColSpan; got != 2 {
		t.Fatalf("colspan = %d, want 2", got)
	}
}

func TestDPDArticleParserPreservesDefinitionQuotesAsAuthored(t *testing.T) {
	article := parseFixture(t, "bien")[0]
	paragraph := article.Sections[0].Paragraphs[0].HTML

	if !strings.Contains(paragraph, "<dfn>'correcta y adecuadamente'</dfn>") {
		t.Fatalf("definition fragment = %q, want authored dfn quotes", paragraph)
	}
	if strings.Contains(paragraph, "\u201c") || strings.Contains(paragraph, "\u201d") {
		t.Fatalf("definition fragment = %q, got synthetic curly quotes", paragraph)
	}
}

func TestDPDArticleParserKeepsExamplesSeparateFromProse(t *testing.T) {
	article := parseFixture(t, "bien")[0]
	paragraph := article.Sections[0].Paragraphs[0].HTML

	if !strings.Contains(paragraph, `<span class="ejemplo">Cierra bien la ventana, por favor</span>`) {
		t.Fatalf("paragraph = %q, want first example span", paragraph)
	}
	if !strings.Contains(paragraph, `<span class="ejemplo">No he dormido bien esta noche</span>`) {
		t.Fatalf("paragraph = %q, want second example span", paragraph)
	}
	if len(article.Sections[0].Paragraphs[0].Inlines) == 0 {
		t.Fatal("paragraph inlines = nil, want structured semantic extraction")
	}
}

func checkBienEmphasisInline(inline model.Inline, sawMention *bool) {
	for _, child := range inline.Children {
		if child.Kind == model.InlineKindMention && (strings.Contains(child.Text, "mejor") || strings.Contains(child.Text, "más bien")) {
			*sawMention = true
		}
	}
}

func checkBienInline(inline model.Inline, sawExample, sawMention, sawGloss, sawReference *bool) {
	switch inline.Kind {
	case model.InlineKindExample:
		if inline.Text == "Cierra bien la ventana, por favor" || inline.Text == "No he dormido bien esta noche" {
			*sawExample = true
		}
	case model.InlineKindEmphasis:
		checkBienEmphasisInline(inline, sawMention)
	case model.InlineKindBracketDefinition:
		if strings.Contains(inline.Text, "correcta y adecuadamente") {
			*sawGloss = true
		}
	case model.InlineKindReference:
		isTargetRef := inline.Text == "6" && inline.Target == "bien#S1590507271213267522"
		if isTargetRef {
			*sawReference = true
		}
	}
}

func TestDPDArticleParserExtractsSemanticInlineKindsFromBien(t *testing.T) {
	article := parseFixture(t, "bien")[0]
	inlines := article.Sections[0].Paragraphs[0].Inlines

	var sawExample, sawMention, sawGloss, sawReference bool
	for _, inline := range inlines {
		checkBienInline(inline, &sawExample, &sawMention, &sawGloss, &sawReference)
	}
	if !sawExample || !sawMention || !sawGloss || !sawReference {
		t.Fatalf("semantic inline extraction incomplete: %#v", inlines)
	}
}

func assertVerCitationMarkers(t *testing.T, inlines []model.Inline) {
	t.Helper()
	if len(inlines) == 0 || inlines[0].Kind != model.InlineKindCitationQuote {
		t.Fatalf("inlines = %#v, want citation quote root", inlines)
	}
	children := inlines[0].Children
	var sawExclusion, sawBib, sawWorkTitle, sawSmallCaps bool
	for _, child := range children {
		switch child.Kind {
		case model.InlineKindExclusion:
			sawExclusion = true
		case model.InlineKindBibliography:
			sawBib = true
			sawWorkTitle, sawSmallCaps = checkBibChildren(child.Children, sawWorkTitle, sawSmallCaps)
		}
	}
	if !sawExclusion || !sawBib || !sawWorkTitle || !sawSmallCaps {
		t.Fatalf("children = %#v", children)
	}
}

func checkBibChildren(grandchildren []model.Inline, sawWorkTitle, sawSmallCaps bool) (bool, bool) {
	for _, gc := range grandchildren {
		if gc.Kind == model.InlineKindWorkTitle {
			sawWorkTitle = true
		}
		if gc.Kind == model.InlineKindSmallCaps {
			sawSmallCaps = true
		}
	}
	return sawWorkTitle, sawSmallCaps
}

func assertDarPatternMarkers(t *testing.T, inlines []model.Inline) {
	t.Helper()
	var sawPattern, sawEditorial, sawCorrection, sawVers bool
	for _, inline := range inlines {
		switch inline.Kind {
		case model.InlineKindPattern:
			sawPattern = true
		case model.InlineKindCitationQuote:
			sawEditorial = checkForEditorial(inline.Children, sawEditorial)
		case model.InlineKindEmphasis:
			sawCorrection, sawVers = checkForCorrectionAndVers(inline.Children, sawCorrection, sawVers)
		}
	}
	if !sawPattern || !sawEditorial || !sawCorrection || !sawVers {
		t.Fatalf("inlines = %#v", inlines)
	}
}

func checkForEditorial(children []model.Inline, sawEditorial bool) bool {
	for _, child := range children {
		if child.Kind == model.InlineKindBracketInterpolation {
			sawEditorial = true
		}
	}
	return sawEditorial
}

func checkForCorrectionAndVers(children []model.Inline, sawCorrection, sawVers bool) (bool, bool) {
	for _, child := range children {
		if child.Kind == model.InlineKindCorrection {
			sawCorrection = true
			for _, gc := range child.Children {
				if gc.Kind == model.InlineKindSmallCaps {
					sawVers = true
				}
			}
		}
	}
	return sawCorrection, sawVers
}

func TestExtractInlinesSupportsVerifiedSemanticMarkersFromVerAndDar(t *testing.T) {
	t.Run("ver citation bibliography and exclusion markers", func(t *testing.T) {
		html := `<span class="cita" n="c"><span class="bolaspa">⊗</span>«Desde atrás vide...» <span class="bib">(González <i>Dios</i> <span class="cbil" title="México">mx</span> 1999)</span></span>`
		assertVerCitationMarkers(t, extractInlines(html))
	})

	t.Run("dar pattern correction and editorial markers", func(t *testing.T) {
		html := `<span class="pattern">«<em><span class="ment">dar</span></em> + gerundio»</span> <span class="cita" n="w">«Las empresas <span class="yy">[…]</span> solo se estaban haciendo cargo» <span class="bib">(<i>Comercio</i><sup>@</sup> <span class="cbil" title="Ecuador">ec</span> 7.10.2016)</span></span> <em><span class="correction">se dio cuenta <span class="vers">de</span> que…</span></em>`
		assertDarPatternMarkers(t, extractInlines(html))
	})
}

func TestDPDArticleParserExtractsNumericReferenceWithoutDuplication(t *testing.T) {
	article := parseFixture(t, "bien")[0]
	paragraph := article.Sections[0].Paragraphs[0].HTML

	if strings.Count(paragraph, "→") != 1 {
		t.Fatalf("paragraph = %q, want exactly one arrow", paragraph)
	}
	if strings.Count(paragraph, `href="bien#S1590507271213267522"`) != 1 {
		t.Fatalf("paragraph = %q, want one reference target", paragraph)
	}
	if strings.Contains(paragraph, "[→ 6]") {
		t.Fatalf("paragraph = %q, parser should preserve source anchor form, not renderer markdown", paragraph)
	}
}

func TestDPDArticleParserPromotesLexicalHeadsIntoTitles(t *testing.T) {
	article := parseFixture(t, "bien")[0]

	tests := []struct {
		index int
		want  string
	}{
		{index: 4, want: "bien que."},
		{index: 5, want: "más bien."},
		{index: 6, want: "si bien."},
	}

	for _, tt := range tests {
		section := article.Sections[tt.index]
		if section.Title != tt.want {
			t.Fatalf("section %d title = %q, want %q", tt.index+1, section.Title, tt.want)
		}
		if len(section.Paragraphs) > 0 && strings.Contains(section.Paragraphs[0].HTML, tt.want) {
			t.Fatalf("section %d paragraph leaked lexical title into body: %q", tt.index+1, section.Paragraphs[0].HTML)
		}
	}
}

func TestExtractInlinesKeepsNestedScaffoldAsFormattingOverride(t *testing.T) {
	inlines := extractInlines(`<em><span class="ment">tilde<span class="nn">2</span></span></em>`)
	if len(inlines) != 1 {
		t.Fatalf("inlines = %#v", inlines)
	}
	if inlines[0].Kind != model.InlineKindEmphasis {
		t.Fatalf("root kind = %q, want emphasis", inlines[0].Kind)
	}
	if len(inlines[0].Children) != 1 || inlines[0].Children[0].Kind != model.InlineKindMention {
		t.Fatalf("children = %#v", inlines[0].Children)
	}
	mention := inlines[0].Children[0]
	if len(mention.Children) != 2 {
		t.Fatalf("mention children = %#v", mention.Children)
	}
	if mention.Children[0].Kind != model.InlineKindText || mention.Children[0].Text != "tilde" {
		t.Fatalf("mention text child = %#v", mention.Children[0])
	}
	if mention.Children[1].Kind != model.InlineKindBracketPronunciation || mention.Children[1].Text != "2" {
		t.Fatalf("mention override child = %#v", mention.Children[1])
	}
}

func TestDPDArticleParserNormalizesHTMLHeadwordToPlainText(t *testing.T) {
	articles := collectArticles(`<entry class="tem" id="tilde2"><header class="tem"><span class="vers">tilde<sup>2</sup></span></header><section class="BLOQUEACEPS"><p n="1n"><span class="enum">1.</span> Variante.</p></section></entry>`, "https://www.rae.es/dpd/tilde")
	if len(articles) != 1 {
		t.Fatalf("articles = %d, want 1", len(articles))
	}
	if articles[0].Lemma != "tilde2" {
		t.Fatalf("lemma = %q, want plain-text tilde2", articles[0].Lemma)
	}
	if strings.Contains(articles[0].Lemma, "<") {
		t.Fatalf("lemma leaked html = %q", articles[0].Lemma)
	}
}

func TestDPDArticleParserIsolatesCitationFromBodyText(t *testing.T) {
	article := parseFixture(t, "bien")[0]

	if !strings.Contains(article.Citation.Text, "Consulta: 10/03/2026") {
		t.Fatalf("citation = %q, want consultation metadata", article.Citation.Text)
	}
	for _, section := range article.Sections {
		for _, paragraph := range section.Paragraphs {
			if strings.Contains(paragraph.HTML, "Consulta:") || strings.Contains(paragraph.HTML, "Real Academia Española") {
				t.Fatalf("section paragraph leaked citation prose: %q", paragraph.HTML)
			}
		}
	}
}

func TestPreserveSemanticSpansEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "input with no HTML tags",
			input: "plain text without any tags at all",
			want:  "plain text without any tags at all",
		},
		{
			name:  "only allowed tags preserved",
			input: `<dfn>definition</dfn> and <em>emphasis</em> and <i>italic</i>`,
			want:  `<dfn>definition</dfn> and <em>emphasis</em> and <i>italic</i>`,
		},
		{
			name:  "allowed span classes preserved",
			input: `<span class="ejemplo">example</span> <span class="ment">mention</span>`,
			want:  `<span class="ejemplo">example</span> <span class="ment">mention</span>`,
		},
		{
			name:  "disallowed tags removed text preserved",
			input: `<div>content inside div</div> <p>paragraph</p>`,
			want:  `content inside div paragraph`,
		},
		{
			name:  "mixed allowed and disallowed tags",
			input: `<div>before</div> <em>emphasized</em> <span class="unknown">removed</span> <dfn>kept</dfn>`,
			want:  `before <em>emphasized</em> removed</span> <dfn>kept</dfn>`,
		},
		{
			name:  "nested tags some allowed some not",
			input: `<div><em>nested emphasis</em></div>`,
			want:  `<em>nested emphasis</em>`,
		},
		{
			name:  "anchor tags preserved as skip prefix",
			input: `text <a href="http://example.com">link</a> more`,
			want:  `text <a href="http://example.com">link</a> more`,
		},
		{
			name:  "table tags preserved as skip prefix",
			input: `<table><tr><td>cell</td></tr></table>`,
			want:  `<table><tr><td>cell</td></tr></table>`,
		},
		{
			name:  "malformed tag no closing angle bracket preserved as-is",
			input: `text <broken tag content`,
			want:  `text <broken tag content`,
		},
		{
			name:  "tags with attributes embf stripped random stripped closing spans kept",
			input: `<span class="embf">bold</span> <span class="random-class">gone</span>`,
			want:  `bold</span> gone</span>`,
		},
		{
			name:  "cita and cbil span prefixes preserved",
			input: `<span class="cita" n="c">citation</span> <span class="cbil" title="mx">mx</span>`,
			want:  `<span class="cita" n="c">citation</span> <span class="cbil" title="mx">mx</span>`,
		},
		{
			name:  "closing span always preserved",
			input: `<span class="ejemplo">text</span>`,
			want:  `<span class="ejemplo">text</span>`,
		},
		{
			name:  "large input with many tags",
			input: strings.Repeat(`<div>text</div> <em>em</em> `, 100),
			want:  strings.Repeat(`text <em>em</em> `, 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preserveSemanticSpans(tt.input)
			if got != tt.want {
				t.Fatalf("preserveSemanticSpans()\n  got:  %q\n  want: %q", got, tt.want)
			}
		})
	}
}

func TestIsChallengePageEdgeCases(t *testing.T) {
	padString := func(prefix string, total int) string {
		if len(prefix) >= total {
			return prefix[:total]
		}
		return prefix + strings.Repeat("x", total-len(prefix))
	}

	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "empty string",
			body: "",
			want: false,
		},
		{
			name: "short content with challenge markers",
			body: "<html>Cloudflare challenge detected</html>",
			want: true,
		},
		{
			name: "short content without challenge markers",
			body: "<html><body>Normal DPD article content</body></html>",
			want: false,
		},
		{
			name: "challenge markers in first 1024 chars",
			body: padString("<html>Cloudflare challenge</html>", 500*1024),
			want: true,
		},
		{
			name: "challenge markers only beyond 1024 chars",
			body: func() string {
				body := strings.Repeat("x", 500*1024)
				// Place markers well beyond 1024 chars
				return body[:2000] + "Cloudflare challenge" + body[2020:]
			}(),
			want: false, // After truncation fix, markers beyond 1024 are not seen
		},
		{
			name: "only cloudflare without challenge",
			body: "Cloudflare CDN network",
			want: false,
		},
		{
			name: "only challenge without cloudflare",
			body: "This is a challenge page",
			want: false,
		},
		{
			name: "markers in mixed case",
			body: "CLOUDFLARE CHALLENGE",
			want: true,
		},
		{
			name: "body exactly 1024 chars with markers",
			body: padString("Cloudflare challenge", 1024),
			want: true,
		},
		{
			name: "non-ASCII content without markers",
			body: "contenido en español: áéíóú ñ ¿¡ «»",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isChallengePage(tt.body)
			if got != tt.want {
				t.Fatalf("isChallengePage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDPDArticleParserDetectsChallengePages(t *testing.T) {
	parser := NewDPDArticleParser()
	_, _, err := parser.Parse(context.Background(), model.SourceDescriptor{Name: "dpd"}, fetch.Document{
		Body: []byte("<html><title>Cloudflare challenge</title><div>challenge</div></html>"),
	})
	if err == nil {
		t.Fatal("Parse() error = nil, want challenge failure")
	}

	problem, ok := model.AsProblem(err)
	if !ok {
		t.Fatalf("Parse() error = %T, want problem", err)
	}
	if problem.Code != model.ProblemCodeDPDExtractFailed {
		t.Fatalf("problem code = %q, want %q", problem.Code, model.ProblemCodeDPDExtractFailed)
	}
}

func TestDPDSignsParsePhase1(t *testing.T) {
	// @ sign
	rawAt := `<sup>@</sup>`
	inlinesAt := extractInlines(rawAt)
	if len(inlinesAt) != 1 || inlinesAt[0].Kind != model.InlineKindDigitalEdition {
		t.Errorf("expected DigitalEdition, got %+v", inlinesAt)
	}

	// + sign
	rawPlus := `<span class="nc">+ infinitivo</span>`
	inlinesPlus := extractInlines(rawPlus)
	if len(inlinesPlus) != 1 || inlinesPlus[0].Kind != model.InlineKindConstructionMarker {
		t.Errorf("expected ConstructionMarker, got %+v", inlinesPlus)
	}
}

func TestDPDSignsParsePhase2(t *testing.T) {
	// Bracket Definition
	rawDef := `<dfn>[una ley]</dfn>`
	inlinesDef := extractInlines(rawDef)
	if len(inlinesDef) != 1 || inlinesDef[0].Kind != model.InlineKindBracketDefinition {
		t.Errorf("expected BracketDefinition, got %+v", inlinesDef)
	}

	// Bracket Pronunciation
	rawPron := `<span class="nn">[alikuóto]</span>`
	inlinesPron := extractInlines(rawPron)
	if len(inlinesPron) != 1 || inlinesPron[0].Kind != model.InlineKindBracketPronunciation {
		t.Errorf("expected BracketPronunciation, got %+v", inlinesPron)
	}

	// Bracket Interpolation
	rawInterp := `<span class="yy">[las feministas]</span>`
	inlinesInterp := extractInlines(rawInterp)
	if len(inlinesInterp) != 1 || inlinesInterp[0].Kind != model.InlineKindBracketInterpolation {
		t.Errorf("expected BracketInterpolation, got %+v", inlinesInterp)
	}
}

func TestDPDSignsParsePhase3Synthetic(t *testing.T) {
	// SYNTHETIC TEST - NO REAL HTML VALIDATION.
	// These patterns are inferred and MUST be updated when real DPD examples are found.
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "agrammatical marker inferred span pattern",
			raw:  `<span class="synthetic-sign">*</span>`,
			want: model.InlineKindAgrammatical,
		},
		{
			name: "hypothetical marker inferred span pattern",
			raw:  `<span class="synthetic-sign">‖</span>`,
			want: model.InlineKindHypothetical,
		},
		{
			name: "phoneme marker inferred plain text pattern",
			raw:  `//`,
			want: model.InlineKindPhoneme,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inlines := extractInlines(tt.raw)
			if len(inlines) != 1 {
				t.Fatalf("extractInlines() len = %d, want 1 (%+v)", len(inlines), inlines)
			}
			if inlines[0].Kind != tt.want {
				t.Fatalf("extractInlines() kind = %q, want %q (%+v)", inlines[0].Kind, tt.want, inlines)
			}
		})
	}
}
