package parse

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

var (
	reEspanolAlDiaContainer = regexp.MustCompile(`(?is)<div\b[^>]*class="[^"]*container pt-8 pb-8 bloque-texto[^"]*"[^>]*>(.*?)</div>\s*<section\b`)
	reEspanolAlDiaTitle     = regexp.MustCompile(`(?is)<h1\b[^>]*class="[^"]*news-title[^"]*"[^>]*>\s*(?:<span>)?(.*?)(?:</span>)?\s*</h1>`)
	reEspanolAlDiaBody      = regexp.MustCompile(`(?is)<div\b[^>]*class="[^"]*col-md-8[^"]*"[^>]*>\s*<div\b[^>]*class="[^"]*pt-4[^"]*"[^>]*>(.*?)</div>\s*</div>`)
	reEspanolAlDiaParagraph = regexp.MustCompile(`(?is)<p\b[^>]*>.*?</p>`)
	reEspanolAlDiaTags      = regexp.MustCompile(`(?is)<[^>]+>`)
)

// EspanolAlDiaParser extracts article-like content from the Español al día surface.
type EspanolAlDiaParser struct{}

// NewEspanolAlDiaParser returns a ready-to-use parser for español-al-día articles.
func NewEspanolAlDiaParser() *EspanolAlDiaParser {
	return &EspanolAlDiaParser{}
}

// Parse extracts the article title and body paragraphs from a fetched page.
func (p *EspanolAlDiaParser) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (Result, []model.Warning, error) {
	_ = ctx
	body := strings.TrimSpace(string(document.Body))
	if body == "" {
		return Result{}, nil, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeArticleExtractFailed,
			Message:  fmt.Sprintf("extract espanol-al-dia article for %q", descriptor.Name),
			Source:   descriptor.Name,
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	container := firstMatchGroup(reEspanolAlDiaContainer.FindStringSubmatch(body), 1)
	if container == "" {
		container = body
	}

	title := normalizeEspanolAlDiaText(firstMatchGroup(reEspanolAlDiaTitle.FindStringSubmatch(container), 1))
	bodyHTML := firstMatchGroup(reEspanolAlDiaBody.FindStringSubmatch(container), 1)
	paragraphMatches := reEspanolAlDiaParagraph.FindAllString(bodyHTML, -1)
	if title == "" || len(paragraphMatches) == 0 {
		return Result{}, nil, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeArticleExtractFailed,
			Message:  fmt.Sprintf("extract espanol-al-dia article for %q", descriptor.Name),
			Source:   descriptor.Name,
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	blocks := make([]ParsedBlock, 0, len(paragraphMatches))
	paragraphs := make([]ParsedParagraph, 0, len(paragraphMatches))
	for _, rawParagraph := range paragraphMatches {
		paragraph := ParsedParagraph{HTML: strings.TrimSpace(rawParagraph)}
		paragraphs = append(paragraphs, paragraph)
		paragraphCopy := paragraph
		blocks = append(blocks, ParsedBlock{Kind: ParsedBlockKindParagraph, Paragraph: &paragraphCopy})
	}

	article := ParsedArticle{
		Dictionary:   "Español al día",
		Lemma:        title,
		CanonicalURL: strings.TrimSpace(document.URL),
		Sections:     []ParsedSection{{Blocks: blocks, Paragraphs: paragraphs}},
	}
	return Result{Articles: []ParsedArticle{article}}, nil, nil
}

func normalizeEspanolAlDiaText(raw string) string {
	text := html.UnescapeString(strings.TrimSpace(raw))
	text = reEspanolAlDiaTags.ReplaceAllString(text, " ")
	text = strings.ReplaceAll(text, "\u00a0", " ")
	return strings.Join(strings.Fields(text), " ")
}
