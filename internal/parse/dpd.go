package parse

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

var (
	reEntry          = regexp.MustCompile(`(?is)<entry\b([^>]*)>(.*?)</entry>`)
	reEntryClass     = regexp.MustCompile(`(?is)\bclass="([^"]+)"`)
	reEntryID        = regexp.MustCompile(`(?is)\bid="([^"]+)"`)
	reEntryHeader    = regexp.MustCompile(`(?is)\bheader="([^"]+)"`)
	reHeader         = regexp.MustCompile(`(?is)<header\b[^>]*class="[^"]*(?:lex|tem)[^"]*"[^>]*>(.*?)</header>`)
	reEdition        = regexp.MustCompile(`(?is)<span class='aviso-edicion'>.*?<span[^>]*>(.*?)</span>.*?</span>`)
	reSectionBlock   = regexp.MustCompile(`(?is)(<p\b([^>]*)>(.*?)</p>)|(<table\b([^>]*)>(.*?)</table>)`)
	reEnum           = regexp.MustCompile(`(?is)<span class="enum">(.*?)</span>`)
	reLexical        = regexp.MustCompile(`(?is)^<span class="embf">(.*?)</span>$`)
	reTags           = regexp.MustCompile(`(?is)<[^>]+>`)
	reCitation       = regexp.MustCompile(`(?is)<p class="o">(.*?)</p>`)
	reReference      = regexp.MustCompile(`(?is)<a\b[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	reEmphasis       = regexp.MustCompile(`(?is)<em\b[^>]*>(.*?)</em>`)
	reAnchorClass    = regexp.MustCompile(`(?is)class="([^"]+)"`)
	reColSpan        = regexp.MustCompile(`(?is)\bcolspan="(\d+)"`)
	reRowSpan        = regexp.MustCompile(`(?is)\browspan="(\d+)"`)
	reTableRow       = regexp.MustCompile(`(?is)<tr\b[^>]*>(.*?)</tr>`)
	reTableCell      = regexp.MustCompile(`(?is)<t[hd]\b([^>]*)>(.*?)</t[hd]>`)
	reTHead          = regexp.MustCompile(`(?is)<thead\b[^>]*>(.*?)</thead>`)
	reTBody          = regexp.MustCompile(`(?is)<tbody\b[^>]*>(.*?)</tbody>`)
	reTHeadCell      = regexp.MustCompile(`(?is)<th\b[^>]*>(.*?)</th>`)
	reClassTokenizer = regexp.MustCompile(`\s+`)
)

var dpdArticleClasses = map[string]bool{
	"lex": true,
	"tem": true,
}

// DPDArticleParser extracts structured articles from DPD HTML pages.
type DPDArticleParser struct{}

// NewDPDArticleParser returns a ready-to-use DPD HTML parser.
func NewDPDArticleParser() *DPDArticleParser {
	return &DPDArticleParser{}
}

// Parse extracts articles from a fetched DPD document, returning structured results and warnings.
func (p *DPDArticleParser) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (Result, []model.Warning, error) {
	_ = ctx
	body := string(document.Body)
	if isChallengePage(body) {
		return Result{}, nil, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDExtractFailed,
			Message:  "DPD response body is a challenge page, not an article",
			Source:   descriptor.Name,
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	query := strings.TrimSpace(descriptor.DisplayName)
	if query == "" {
		query = strings.TrimSpace(descriptor.Name)
	}

	articles := collectArticles(body, document.URL)
	if len(articles) == 0 {
		return Result{}, nil, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDNotFound,
			Message:  fmt.Sprintf("DPD entry not found for %q", query),
			Source:   descriptor.Name,
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	warnings := []model.Warning{accessWarning(descriptor.Name)}
	return Result{Articles: articles}, warnings, nil
}

func collectArticles(body, canonicalURL string) []ParsedArticle {
	edition := extractFirstText(reEdition, body)
	citationHTML := extractFirstMatch(reCitation, body)
	citation := ParsedCitation{HTML: citationHTML, Text: citationText(citationHTML)}
	articles := make([]ParsedArticle, 0, 4)

	for _, match := range reEntry.FindAllStringSubmatch(body, -1) {
		attrs := match[1]
		entryHTML := match[2]
		if !isRelevantArticleEntry(attrs) {
			continue
		}

		sections := parseArticleSections(entryHTML)
		if len(sections) == 0 {
			continue
		}

		entryID := extractAttribute(reEntryID, attrs)
		header := normalizeInlinePlainText(extractFirstMatch(reHeader, entryHTML))
		if header == "" {
			header = normalizeInlinePlainText(extractAttribute(reEntryHeader, attrs))
		}
		if header == "" {
			header = cleanText(entryID)
		}

		articles = append(articles, ParsedArticle{
			Dictionary:   "Diccionario panhispánico de dudas",
			Edition:      edition,
			EntryID:      entryID,
			Lemma:        header,
			CanonicalURL: canonicalURL,
			Sections:     sections,
			Citation:     citation,
		})
	}

	return articles
}

func isRelevantArticleEntry(attrs string) bool {
	classAttr := extractAttribute(reEntryClass, attrs)
	if classAttr == "" {
		return false
	}
	for _, className := range reClassTokenizer.Split(strings.TrimSpace(classAttr), -1) {
		if dpdArticleClasses[strings.ToLower(strings.TrimSpace(className))] {
			return true
		}
	}
	return false
}

// sectionBlockAttrs unpacks a reSectionBlock submatch into tag name, attrs, and inner HTML.
// Returns ("", "", "") when the match group is empty or the inner HTML is blank.
func sectionBlockAttrs(match []string) (tag, attrs, innerHTML string) {
	if match[4] != "" {
		innerHTML = strings.TrimSpace(match[6])
		if innerHTML == "" {
			return "", "", ""
		}
		return "table", match[5], innerHTML
	}
	innerHTML = strings.TrimSpace(match[3])
	if innerHTML == "" {
		return "", "", ""
	}
	return "p", match[2], innerHTML
}

// buildLabeledSection constructs a ParsedSection from a labeled paragraph match.
func buildLabeledSection(label, cleanedHTML string) ParsedSection {
	section := ParsedSection{Label: html.UnescapeString(strings.TrimSpace(label))}
	if title, ok := lexicalTitle(cleanedHTML); ok {
		section.Title = title
		return section
	}
	paragraph := newParsedParagraph(cleanedHTML)
	if paragraph.HTML != "" {
		section.Blocks = []ParsedBlock{{Kind: ParsedBlockKindParagraph, Paragraph: &paragraph}}
		section.Paragraphs = []ParsedParagraph{paragraph}
	}
	return section
}

// routeSection inserts section into the right slot (nested child or new top-level).
// Returns false when a nested section is encountered with no parent, which signals a parse failure.
func routeSection(section ParsedSection, attrs string, parsedSections *[]ParsedSection, currentTop, currentNested **ParsedSection) bool {
	if strings.Contains(attrs, `n="2l"`) || isNestedLabel(section.Label) {
		if *currentTop == nil {
			return false
		}
		(*currentTop).Children = append((*currentTop).Children, section)
		*currentNested = &(*currentTop).Children[len((*currentTop).Children)-1]
		return true
	}
	*parsedSections = append(*parsedSections, section)
	*currentTop = &(*parsedSections)[len(*parsedSections)-1]
	*currentNested = nil
	return true
}

func parseArticleSections(articleHTML string) []ParsedSection {
	parsedSections := make([]ParsedSection, 0, 8)
	var currentTop *ParsedSection
	var currentNested *ParsedSection

	for _, match := range reSectionBlock.FindAllStringSubmatch(articleHTML, -1) {
		tag, attrs, innerHTML := sectionBlockAttrs(match)
		if innerHTML == "" {
			continue
		}

		if tag == "table" {
			appendBlock(&currentTop, &currentNested, ParsedBlock{Kind: ParsedBlockKindTable, Table: parseTable(innerHTML)})
			continue
		}

		label := extractFirstText(reEnum, innerHTML)
		cleanedHTML := strings.TrimSpace(reEnum.ReplaceAllString(innerHTML, ""))
		if label == "" {
			paragraph := newParsedParagraph(cleanedHTML)
			if paragraph.HTML == "" {
				continue
			}
			appendBlock(&currentTop, &currentNested, ParsedBlock{Kind: ParsedBlockKindParagraph, Paragraph: &paragraph})
			continue
		}

		section := buildLabeledSection(label, cleanedHTML)
		if !routeSection(section, attrs, &parsedSections, &currentTop, &currentNested) {
			return nil
		}
	}

	return parsedSections
}

func appendBlock(currentTop, currentNested **ParsedSection, block ParsedBlock) {
	target := *currentNested
	if target == nil {
		target = *currentTop
	}
	if target == nil {
		return
	}
	target.Blocks = append(target.Blocks, block)
	if block.Kind == ParsedBlockKindParagraph && block.Paragraph != nil {
		target.Paragraphs = append(target.Paragraphs, *block.Paragraph)
	}
}

func parseTable(innerHTML string) *ParsedTable {
	table := &ParsedTable{}
	theadHTML := extractFirstMatch(reTHead, innerHTML)
	if theadHTML != "" {
		table.Headers = parseTableRows(theadHTML, true)
	}
	tbodyHTML := extractFirstMatch(reTBody, innerHTML)
	if tbodyHTML != "" {
		table.Rows = parseTableRows(tbodyHTML, false)
	} else {
		table.Rows = parseTableRows(innerHTML, false)
		if len(table.Headers) == 0 && len(table.Rows) > 0 && tableRowLooksLikeHeader(innerHTML) {
			table.Headers = append(table.Headers, table.Rows[0])
			table.Rows = table.Rows[1:]
		}
	}
	return table
}

func parseTableRows(raw string, headerOnly bool) []ParsedTableRow {
	rows := make([]ParsedTableRow, 0, 4)
	for _, rowMatch := range reTableRow.FindAllStringSubmatch(raw, -1) {
		rowHTML := rowMatch[1]
		if headerOnly && len(reTHeadCell.FindAllStringSubmatch(rowHTML, -1)) == 0 {
			continue
		}
		cells := make([]ParsedTableCell, 0, 4)
		for _, cellMatch := range reTableCell.FindAllStringSubmatch(rowHTML, -1) {
			cells = append(cells, cleanTableCell(cellMatch[1], cellMatch[2]))
		}
		if len(cells) == 0 {
			continue
		}
		rows = append(rows, ParsedTableRow{Cells: cells})
	}
	return rows
}

func tableRowLooksLikeHeader(raw string) bool {
	firstRow := reTableRow.FindString(raw)
	if firstRow == "" {
		return false
	}
	return len(reTHeadCell.FindAllStringSubmatch(firstRow, -1)) > 0
}

func cleanTableCell(attrs, raw string) ParsedTableCell {
	raw = normalizeHTMLParagraph(raw)
	return ParsedTableCell{
		HTML:    raw,
		Inlines: extractInlines(raw),
		ColSpan: positiveSpan(extractAttribute(reColSpan, attrs)),
		RowSpan: positiveSpan(extractAttribute(reRowSpan, attrs)),
	}
}

func positiveSpan(raw string) int {
	if raw == "" {
		return 0
	}
	if raw == "1" {
		return 1
	}
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0
		}
	}
	if raw == "0" {
		return 0
	}
	value := 0
	for _, ch := range raw {
		value = value*10 + int(ch-'0')
	}
	if value < 1 {
		return 0
	}
	return value
}

func extractAttribute(re *regexp.Regexp, attrs string) string {
	match := re.FindStringSubmatch(attrs)
	if len(match) < 2 {
		return ""
	}
	return html.UnescapeString(strings.TrimSpace(match[1]))
}

const (
	challengePageSnippetLimit = 1024
	exclusionGlyph            = "\u2297" // ⊗
	digitalEditionGlyph       = "@"
	constructionMarkerGlyph   = "+"

	// WARNING: Speculative signs - no real HTML validation found in DPD articles.
	// Patterns are inferred from validated sign handling and MUST be updated when
	// real examples are discovered.
	agrammaticalGlyph = "*"
	hypotheticalGlyph = "‖" // \u2016
	phonemeGlyph      = "//"
)

// ARCHIVED SIGNS (not implemented):
// < (etymology) and > (transformation) remain excluded because of HTML tag
// collision risk. See testdata/dpd-signs-analysis/SIGN_ANALYSIS.md.

func isChallengePage(body string) bool {
	snippet := body
	if len(snippet) > challengePageSnippetLimit {
		snippet = snippet[:challengePageSnippetLimit]
	}
	lower := strings.ToLower(snippet)
	return strings.Contains(lower, "cloudflare") && strings.Contains(lower, "challenge")
}

func isNestedLabel(label string) bool {
	trimmed := strings.TrimSpace(label)
	return strings.HasSuffix(trimmed, ")") && len(trimmed) <= 3
}

func normalizeHTMLParagraph(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, `<br>`, "\n")
	raw = strings.ReplaceAll(raw, `<br/>`, "\n")
	raw = strings.ReplaceAll(raw, `<br />`, "\n")
	raw = preserveSemanticSpans(raw)
	return strings.TrimSpace(raw)
}

func newParsedParagraph(raw string) ParsedParagraph {
	normalized := normalizeHTMLParagraph(raw)
	return ParsedParagraph{
		HTML:    normalized,
		Inlines: extractInlines(normalized),
	}
}

// inlineFrame is a parse-stack entry for a currently-open inline element.
type inlineFrame struct {
	tag    string
	inline model.Inline
}

// inlineParser holds the mutable state threaded through extractInlines.
type inlineParser struct {
	root  []model.Inline
	stack []inlineFrame
}

// appendNode adds a completed inline to the innermost open frame, or to the root list.
func (p *inlineParser) appendNode(inline model.Inline) {
	if len(p.stack) == 0 {
		p.root = append(p.root, inline)
		return
	}
	p.stack[len(p.stack)-1].inline.Children = append(p.stack[len(p.stack)-1].inline.Children, inline)
}

// resolveFrameText fills the Text field of a completed frame according to its kind.
func resolveFrameText(frame *inlineFrame) {
	if len(frame.inline.Children) > 0 && frame.inline.Kind == model.InlineKindReference {
		frame.inline.Text = cleanText(renderInlineChildrenText(frame.inline.Children))
	}
	if frame.inline.Kind != model.InlineKindReference && frame.inline.Kind != model.InlineKindExclusion {
		frame.inline.Text = cleanText(renderInlineChildrenText(frame.inline.Children))
	}
}

// drainUntil pops frames off the stack until the frame matching closeTag is emitted.
// When closeTag is empty it drains the entire stack (used for end-of-input cleanup).
func (p *inlineParser) drainUntil(closeTag string) {
	for len(p.stack) > 0 {
		frame := p.stack[len(p.stack)-1]
		p.stack = p.stack[:len(p.stack)-1]
		resolveFrameText(&frame)
		p.appendNode(frame.inline)
		if closeTag != "" && frame.tag == closeTag {
			break
		}
	}
}

// handleCloseTag processes a closing-tag token from the input stream.
func (p *inlineParser) handleCloseTag(lower string) {
	closeTag := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lower, "</"), ">"))
	p.drainUntil(closeTag)
}

// advanceTextBefore emits a text node for any plain content before the next '<' and
// returns the slice of raw starting at that '<'. When there is no '<', it emits the
// whole remaining string, appends it, and returns "".
func (p *inlineParser) advanceTextBefore(raw string) string {
	nextTag := strings.Index(raw, "<")
	if nextTag == -1 {
		appendPlainOrSpeculativeInline(p, raw)
		return ""
	}
	if nextTag > 0 {
		appendPlainOrSpeculativeInline(p, raw[:nextTag])
		return raw[nextTag:]
	}
	return raw
}

func extractInlines(raw string) []model.Inline {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	p := &inlineParser{}

	for len(raw) > 0 {
		// Emit any text before the next '<', or consume no-tag trailing text.
		if !strings.HasPrefix(raw, "<") {
			raw = p.advanceTextBefore(raw)
			continue
		}

		end := strings.Index(raw, ">")
		if end == -1 {
			// Malformed: tag opened but never closed — treat remainder as text.
			appendPlainOrSpeculativeInline(p, raw)
			break
		}

		tagToken := raw[:end+1]
		raw = raw[end+1:]
		lower := strings.ToLower(tagToken)

		if strings.HasPrefix(lower, "</") {
			p.handleCloseTag(lower)
			continue
		}

		if inline, closeTag, ok := parseSupportedOpenTag(tagToken, raw); ok {
			p.stack = append(p.stack, inlineFrame{tag: closeTag, inline: inline})
		}
	}

	// Flush any frames that were never closed by a matching close tag.
	p.drainUntil("")

	return mergeTextInlines(p.root)
}

func mergeTextInlines(inlines []model.Inline) []model.Inline {
	merged := make([]model.Inline, 0, len(inlines))
	for _, inline := range inlines {
		if inline.Kind == model.InlineKindText && len(merged) > 0 && merged[len(merged)-1].Kind == model.InlineKindText {
			merged[len(merged)-1].Text += inline.Text
			continue
		}
		merged = append(merged, inline)
	}
	return merged
}

func parseSupportedOpenTag(tag, remaining string) (model.Inline, string, bool) {
	lower := strings.ToLower(tag)
	inline := model.Inline{}
	switch {
	case strings.HasPrefix(lower, "<span"):
		if classParts := reAnchorClass.FindStringSubmatch(tag); len(classParts) == 2 {
			inline.Variant = cleanText(classParts[1])
		}

		// WARNING: Speculative signs - no real HTML validation exists in DPD
		// fixtures yet. This is a best-guess lookahead based on the current span
		// container shape. Update when real examples are discovered.
		if kind, ok := inferSpeculativeSpanKind(remaining); ok {
			inline.Kind = kind
			return inline, "span", true
		}

		switch inline.Variant {
		case "ejemplo":
			inline.Kind = model.InlineKindExample
		case "ment":
			inline.Kind = model.InlineKindMention
		case "embf":
			inline.Kind = model.InlineKindLexicalHeading
		case "cita":
			inline.Kind = model.InlineKindCitationQuote
		case "bib":
			inline.Kind = model.InlineKindBibliography
		case "vers", "cbil":
			inline.Kind = model.InlineKindSmallCaps
		case "yy":
			inline.Kind = model.InlineKindBracketInterpolation
		case "bolaspa":
			inline.Kind = model.InlineKindExclusion
		case "nn":
			inline.Kind = model.InlineKindBracketPronunciation
		case "nc":
			inline.Kind = model.InlineKindConstructionMarker
		case "pattern":
			inline.Kind = model.InlineKindPattern
		case "correction":
			inline.Kind = model.InlineKindCorrection
		default:
			return model.Inline{}, "", false
		}
		return inline, "span", true
	case strings.HasPrefix(lower, "<dfn"):
		inline.Kind = model.InlineKindBracketDefinition
		return inline, "dfn", true
	case strings.HasPrefix(lower, "<em"):
		inline.Kind = model.InlineKindEmphasis
		return inline, "em", true
	case strings.HasPrefix(lower, "<i"):
		inline.Kind = model.InlineKindWorkTitle
		return inline, "i", true
	case strings.HasPrefix(lower, "<sup"):
		if cleanInlineSegment(previewInnerHTML(remaining, "sup")) == digitalEditionGlyph {
			inline.Kind = model.InlineKindDigitalEdition
		} else {
			inline.Kind = model.InlineKindText
		}
		return inline, "sup", true
	case strings.HasPrefix(lower, "<a "):
		parts := reReference.FindStringSubmatch(tag + "</a>")
		if len(parts) >= 2 {
			inline.Target = html.UnescapeString(parts[1])
		}
		if classParts := reAnchorClass.FindStringSubmatch(tag); len(classParts) == 2 {
			inline.Variant = cleanText(classParts[1])
		}
		inline.Kind = model.InlineKindReference
		return inline, "a", true
	default:
		return model.Inline{}, "", false
	}
}

func inferSpeculativeSpanKind(remaining string) (string, bool) {
	preview := cleanInlineSegment(previewInnerHTML(remaining, "span"))
	switch preview {
	case agrammaticalGlyph:
		return model.InlineKindAgrammatical, true
	case hypotheticalGlyph:
		return model.InlineKindHypothetical, true
	case phonemeGlyph:
		return model.InlineKindPhoneme, true
	default:
		return "", false
	}
}

func previewInnerHTML(raw, closeTag string) string {
	lower := strings.ToLower(raw)
	needle := "</" + closeTag + ">"
	end := strings.Index(lower, needle)
	if end == -1 {
		return ""
	}
	return raw[:end]
}

func renderInlineChildrenText(inlines []model.Inline) string {
	var builder strings.Builder
	for _, inline := range inlines {
		text := inline.Text
		if len(inline.Children) > 0 {
			text = renderInlineChildrenText(inline.Children)
		}
		if text == "" {
			continue
		}
		if builder.Len() > 0 && needsInlineTextSpace(builder.String(), text) {
			builder.WriteString(" ")
		}
		builder.WriteString(text)
	}
	return strings.TrimSpace(builder.String())
}

func needsInlineTextSpace(current, next string) bool {
	if current == "" || next == "" {
		return false
	}
	last := rune(current[len(current)-1])
	first := rune(next[0])
	if (unicode.IsLetter(last) && unicode.IsDigit(first)) || (unicode.IsDigit(last) && unicode.IsLetter(first)) {
		return false
	}
	if strings.ContainsRune("([{«", last) || strings.ContainsRune(")]},.;:!?»", first) {
		return false
	}
	return true
}

func normalizeInlinePlainText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	normalized := normalizeHTMLParagraph(raw)
	inlines := extractInlines(normalized)
	if len(inlines) > 0 {
		return cleanText(renderInlineChildrenText(inlines))
	}
	return cleanText(reTags.ReplaceAllString(normalized, ""))
}

func appendPlainOrSpeculativeInline(p *inlineParser, raw string) {
	text := cleanInlineSegment(raw)
	if text == "" {
		return
	}

	// WARNING: Speculative phoneme handling has no real DPD HTML validation.
	// Only isolated plain-text // is upgraded; broader patterns must wait for
	// real examples.
	if text == phonemeGlyph {
		p.appendNode(model.Inline{Kind: model.InlineKindPhoneme, Text: text})
		return
	}

	p.appendNode(model.Inline{Kind: model.InlineKindText, Text: text})
}

var semanticSpanAllowed = map[string]bool{
	"<dfn>":                       true,
	"</dfn>":                      true,
	"<em>":                        true,
	"</em>":                       true,
	"<i>":                         true,
	"</i>":                        true,
	"<sup>":                       true,
	"</sup>":                      true,
	"<span class=\"ejemplo\">":    true,
	"<span class=\"ment\">":       true,
	"<span class=\"bib\">":        true,
	"<span class=\"vers\">":       true,
	"<span class=\"yy\">":         true,
	"<span class=\"bolaspa\">":    true,
	"<span class=\"nn\">":         true,
	"<span class=\"nc\">":         true,
	"<span class=\"pattern\">":    true,
	"<span class=\"correction\">": true,
	"</span>":                     true,
}

func isSemanticSkipPrefix(lower string) bool {
	return strings.HasPrefix(lower, `<a `) || strings.HasPrefix(lower, `</a`) ||
		strings.HasPrefix(lower, `<table`) || strings.HasPrefix(lower, `</table`) ||
		strings.HasPrefix(lower, `<tr`) || strings.HasPrefix(lower, `</tr`) ||
		strings.HasPrefix(lower, `<td`) || strings.HasPrefix(lower, `</td`) ||
		strings.HasPrefix(lower, `<th`) || strings.HasPrefix(lower, `</th`) ||
		strings.HasPrefix(lower, `<thead`) || strings.HasPrefix(lower, `</thead`) ||
		strings.HasPrefix(lower, `<tbody`) || strings.HasPrefix(lower, `</tbody`)
}

func isAllowedTag(lower string) bool {
	return semanticSpanAllowed[lower] ||
		strings.HasPrefix(lower, `<span class="cita"`) ||
		strings.HasPrefix(lower, `<span class="cbil"`)
}

func preserveSemanticSpans(raw string) string {
	// Fast path: no tags at all.
	if !strings.Contains(raw, "<") {
		return raw
	}

	var builder strings.Builder
	builder.Grow(len(raw))
	pos := 0

	for pos < len(raw) {
		nextTag := strings.Index(raw[pos:], "<")
		if nextTag == -1 {
			builder.WriteString(raw[pos:])
			break
		}
		// Copy text before the tag.
		builder.WriteString(raw[pos : pos+nextTag])
		tagStart := pos + nextTag

		endTag := strings.Index(raw[tagStart:], ">")
		if endTag == -1 {
			// Malformed: no closing '>'. Copy the rest as-is.
			builder.WriteString(raw[tagStart:])
			break
		}
		tag := raw[tagStart : tagStart+endTag+1]
		lower := strings.ToLower(tag)

		if isSemanticSkipPrefix(lower) || isAllowedTag(lower) {
			builder.WriteString(tag)
		}
		// else: drop the tag (disallowed)

		pos = tagStart + endTag + 1
	}

	return builder.String()
}

func cleanText(raw string) string {
	text := html.UnescapeString(raw)
	text = strings.ReplaceAll(text, "\u200d", "")
	text = strings.ReplaceAll(text, exclusionGlyph, exclusionGlyph)                   //nolint:gocritic // dupArg: intentional normalization of ⊗ variants from HTML entities
	text = strings.ReplaceAll(text, digitalEditionGlyph, digitalEditionGlyph)         //nolint:gocritic // Preserve @ sign
	text = strings.ReplaceAll(text, constructionMarkerGlyph, constructionMarkerGlyph) //nolint:gocritic // Preserve + sign
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func cleanInlineSegment(raw string) string {
	text := html.UnescapeString(reTags.ReplaceAllString(raw, ""))
	text = strings.ReplaceAll(text, "\u200d", "")
	text = strings.ReplaceAll(text, exclusionGlyph, exclusionGlyph)                   //nolint:gocritic // dupArg: intentional normalization of ⊗ variants from HTML entities
	text = strings.ReplaceAll(text, digitalEditionGlyph, digitalEditionGlyph)         //nolint:gocritic // Preserve @ sign
	text = strings.ReplaceAll(text, constructionMarkerGlyph, constructionMarkerGlyph) //nolint:gocritic // Preserve + sign
	// WARNING: Speculative signs - no real HTML validation. Patterns are inferred
	// and MUST be updated when real DPD examples are discovered.
	text = strings.ReplaceAll(text, agrammaticalGlyph, agrammaticalGlyph) //nolint:gocritic // Preserve * sign (SPECULATIVE)
	text = strings.ReplaceAll(text, hypotheticalGlyph, hypotheticalGlyph) //nolint:gocritic // Preserve ‖ sign (SPECULATIVE)
	text = strings.ReplaceAll(text, phonemeGlyph, phonemeGlyph)           //nolint:gocritic // Preserve // sign (SPECULATIVE)
	leading := strings.HasPrefix(text, " ") || strings.HasPrefix(text, "\n") || strings.HasPrefix(text, "\t")
	trailing := strings.HasSuffix(text, " ") || strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\t")
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return ""
	}
	if leading {
		text = " " + text
	}
	if trailing {
		text += " "
	}
	return text
}

func extractFirstText(re *regexp.Regexp, body string) string {
	match := re.FindStringSubmatch(body)
	if len(match) < 2 {
		return ""
	}
	return cleanText(match[1])
}

func extractFirstMatch(re *regexp.Regexp, body string) string {
	match := re.FindStringSubmatch(body)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func lexicalTitle(raw string) (string, bool) {
	match := reLexical.FindStringSubmatch(strings.TrimSpace(raw))
	if len(match) < 2 {
		return "", false
	}

	return normalizeInlinePlainText(match[1]), true
}

func accessWarning(source string) model.Warning {
	return model.Warning{
		Code:    "dpd_access_profile",
		Source:  source,
		Message: "validated access method: direct GET /dpd/<term> with browser-like User-Agent reaches article HTML; low-profile/no-UA requests may trigger Cloudflare challenge pages; /srv/keys is useful for entry discovery only and its remote JSON decoding belongs in the dedicated search parse layer; go-rae is not a direct DPD blueprint",
	}
}

func citationText(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = reReference.ReplaceAllString(raw, `$2`)
	raw = reEmphasis.ReplaceAllString(raw, `$1`)
	raw = strings.ReplaceAll(raw, `<i>`, "")
	raw = strings.ReplaceAll(raw, `</i>`, "")
	raw = strings.ReplaceAll(raw, `<br>`, " ")
	raw = strings.ReplaceAll(raw, `<br/>`, " ")
	raw = strings.ReplaceAll(raw, `<br />`, " ")
	raw = reTags.ReplaceAllString(raw, "")
	return cleanText(raw)
}
