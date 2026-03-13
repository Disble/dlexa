package parse

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/gentleman-programming/dlexa/internal/fetch"
	"github.com/gentleman-programming/dlexa/internal/model"
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

type DPDArticleParser struct{}

func NewDPDArticleParser() *DPDArticleParser {
	return &DPDArticleParser{}
}

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

func parseArticleSections(articleHTML string) []ParsedSection {
	parsedSections := make([]ParsedSection, 0, 8)
	var currentTop *ParsedSection
	var currentNested *ParsedSection

	for _, match := range reSectionBlock.FindAllStringSubmatch(articleHTML, -1) {
		tag := "p"
		attrs := match[2]
		innerHTML := strings.TrimSpace(match[3])
		if match[4] != "" {
			tag = "table"
			attrs = match[5]
			innerHTML = strings.TrimSpace(match[6])
		}
		if innerHTML == "" {
			continue
		}

		if tag == "table" {
			block := ParsedBlock{Kind: ParsedBlockKindTable, Table: parseTable(innerHTML)}
			appendBlock(&currentTop, &currentNested, block)
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

		section := ParsedSection{Label: html.UnescapeString(strings.TrimSpace(label))}
		if title, ok := lexicalTitle(cleanedHTML); ok {
			section.Title = title
		} else {
			paragraph := newParsedParagraph(cleanedHTML)
			if paragraph.HTML != "" {
				section.Blocks = []ParsedBlock{{Kind: ParsedBlockKindParagraph, Paragraph: &paragraph}}
				section.Paragraphs = []ParsedParagraph{paragraph}
			}
		}

		if strings.Contains(attrs, `n="2l"`) || isNestedLabel(section.Label) {
			if currentTop == nil {
				return nil
			}
			currentTop.Children = append(currentTop.Children, section)
			currentNested = &currentTop.Children[len(currentTop.Children)-1]
			continue
		}

		parsedSections = append(parsedSections, section)
		currentTop = &parsedSections[len(parsedSections)-1]
		currentNested = nil
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

func cleanTableCell(attrs string, raw string) ParsedTableCell {
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

func isChallengePage(body string) bool {
	lower := strings.ToLower(body)
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

func extractInlines(raw string) []model.Inline {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	type inlineFrame struct {
		tag    string
		inline model.Inline
	}

	var root []model.Inline
	var stack []inlineFrame

	appendNode := func(inline model.Inline) {
		if len(stack) == 0 {
			root = append(root, inline)
			return
		}
		stack[len(stack)-1].inline.Children = append(stack[len(stack)-1].inline.Children, inline)
	}

	for len(raw) > 0 {
		nextTag := strings.Index(raw, "<")
		if nextTag == -1 {
			if text := cleanInlineSegment(raw); text != "" {
				appendNode(model.Inline{Kind: model.InlineKindText, Text: text})
			}
			break
		}
		if nextTag > 0 {
			if text := cleanInlineSegment(raw[:nextTag]); text != "" {
				appendNode(model.Inline{Kind: model.InlineKindText, Text: text})
			}
			raw = raw[nextTag:]
			continue
		}

		end := strings.Index(raw, ">")
		if end == -1 {
			if text := cleanInlineSegment(raw); text != "" {
				appendNode(model.Inline{Kind: model.InlineKindText, Text: text})
			}
			break
		}

		tagToken := raw[:end+1]
		raw = raw[end+1:]
		lower := strings.ToLower(tagToken)
		if strings.HasPrefix(lower, "</") {
			closeTag := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lower, "</"), ">"))
			for len(stack) > 0 {
				frame := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if len(frame.inline.Children) > 0 && frame.inline.Kind == model.InlineKindReference {
					frame.inline.Text = cleanText(renderInlineChildrenText(frame.inline.Children))
				}
				if frame.inline.Kind != model.InlineKindReference && frame.inline.Kind != model.InlineKindExclusion {
					frame.inline.Text = cleanText(renderInlineChildrenText(frame.inline.Children))
				}
				appendNode(frame.inline)
				if frame.tag == closeTag {
					break
				}
			}
			continue
		}

		if inline, closeTag, ok := parseSupportedOpenTag(tagToken); ok {
			stack = append(stack, inlineFrame{tag: closeTag, inline: inline})
		}
	}

	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(frame.inline.Children) > 0 && frame.inline.Kind == model.InlineKindReference {
			frame.inline.Text = cleanText(renderInlineChildrenText(frame.inline.Children))
		}
		if frame.inline.Kind != model.InlineKindReference && frame.inline.Kind != model.InlineKindExclusion {
			frame.inline.Text = cleanText(renderInlineChildrenText(frame.inline.Children))
		}
		appendNode(frame.inline)
	}

	return mergeTextInlines(root)
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

func parseSupportedOpenTag(tag string) (model.Inline, string, bool) {
	lower := strings.ToLower(tag)
	inline := model.Inline{}
	switch {
	case strings.HasPrefix(lower, "<span"):
		classParts := reAnchorClass.FindStringSubmatch(tag)
		if len(classParts) != 2 {
			return model.Inline{}, "", false
		}
		className := cleanText(classParts[1])
		inline.Variant = className
		switch className {
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
			inline.Kind = model.InlineKindEditorial
		case "bolaspa":
			inline.Kind = model.InlineKindExclusion
		case "nn", "nc":
			inline.Kind = model.InlineKindScaffold
		case "pattern":
			inline.Kind = model.InlineKindPattern
		case "correction":
			inline.Kind = model.InlineKindCorrection
		default:
			return model.Inline{}, "", false
		}
		return inline, "span", true
	case strings.HasPrefix(lower, "<dfn"):
		inline.Kind = model.InlineKindGloss
		return inline, "dfn", true
	case strings.HasPrefix(lower, "<em"):
		inline.Kind = model.InlineKindEmphasis
		return inline, "em", true
	case strings.HasPrefix(lower, "<i"):
		inline.Kind = model.InlineKindWorkTitle
		return inline, "i", true
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

func preserveSemanticSpans(raw string) string {
	parts := reTags.FindAllString(raw, -1)
	if len(parts) == 0 {
		return raw
	}

	allowed := map[string]bool{
		"<dfn>":                       true,
		"</dfn>":                      true,
		"<em>":                        true,
		"</em>":                       true,
		"<i>":                         true,
		"</i>":                        true,
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

	for _, tag := range parts {
		lower := strings.ToLower(tag)
		if strings.HasPrefix(lower, `<a `) || strings.HasPrefix(lower, `</a`) {
			continue
		}
		if strings.HasPrefix(lower, `<table`) || strings.HasPrefix(lower, `</table`) || strings.HasPrefix(lower, `<tr`) || strings.HasPrefix(lower, `</tr`) || strings.HasPrefix(lower, `<td`) || strings.HasPrefix(lower, `</td`) || strings.HasPrefix(lower, `<th`) || strings.HasPrefix(lower, `</th`) || strings.HasPrefix(lower, `<thead`) || strings.HasPrefix(lower, `</thead`) || strings.HasPrefix(lower, `<tbody`) || strings.HasPrefix(lower, `</tbody`) {
			continue
		}
		if allowed[lower] || strings.HasPrefix(lower, `<span class="cita"`) || strings.HasPrefix(lower, `<span class="cbil"`) {
			continue
		}
		raw = strings.ReplaceAll(raw, tag, "")
	}

	return raw
}

func cleanText(raw string) string {
	text := html.UnescapeString(raw)
	text = strings.ReplaceAll(text, "\u200d", "")
	text = strings.ReplaceAll(text, "⊗", "⊗")
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func cleanInlineSegment(raw string) string {
	text := html.UnescapeString(reTags.ReplaceAllString(raw, ""))
	text = strings.ReplaceAll(text, "\u200d", "")
	text = strings.ReplaceAll(text, "⊗", "⊗")
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
		text = text + " "
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
		Message: "validated access method: direct GET /dpd/<term> with browser-like User-Agent reaches article HTML; low-profile/no-UA requests may trigger Cloudflare challenge pages; /srv/keys is not useful here; go-rae is not a direct DPD blueprint",
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
