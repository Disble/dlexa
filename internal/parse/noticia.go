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
	reNoticiaTitle     = regexp.MustCompile(`(?is)<h1\b[^>]*class="[^"]*news-title[^"]*"[^>]*>\s*(?:<span>)?(.*?)(?:</span>)?\s*</h1>`)
	reNoticiaBody      = regexp.MustCompile(`(?is)<div\b[^>]*class="[^"]*bloque-texto[^"]*"[^>]*>(.*?)</div>`)
	reNoticiaParagraph = regexp.MustCompile(`(?is)<p\b[^>]*>.*?</p>`)
	reNoticiaTags      = regexp.MustCompile(`(?is)<[^>]+>`)
)

// NoticiaParser extracts FAQ-style article content from the noticia surface.
type NoticiaParser struct{}

// NewNoticiaParser returns a ready-to-use parser for noticia pages.
func NewNoticiaParser() *NoticiaParser {
	return &NoticiaParser{}
}

// Parse extracts the article title and body paragraphs from a fetched noticia page.
func (p *NoticiaParser) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (Result, []model.Warning, error) {
	_ = ctx
	body := strings.TrimSpace(string(document.Body))
	if body == "" {
		return Result{}, nil, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeArticleExtractFailed,
			Message:  fmt.Sprintf("extract noticia article for %q", descriptor.Name),
			Source:   descriptor.Name,
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	title := normalizeNoticiaText(firstMatchGroup(reNoticiaTitle.FindStringSubmatch(body), 1))
	bodyHTML := firstMatchGroup(reNoticiaBody.FindStringSubmatch(body), 1)
	paragraphMatches := reNoticiaParagraph.FindAllString(bodyHTML, -1)
	if title == "" || len(paragraphMatches) == 0 {
		return Result{}, nil, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeArticleExtractFailed,
			Message:  fmt.Sprintf("extract noticia article for %q", descriptor.Name),
			Source:   descriptor.Name,
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	blocks := make([]ParsedBlock, 0, len(paragraphMatches))
	paragraphs := make([]ParsedParagraph, 0, len(paragraphMatches))
	for _, rawParagraph := range paragraphMatches {
		trimmed := strings.TrimSpace(rawParagraph)
		if trimmed == "" || trimmed == "<p>&nbsp;</p>" {
			continue
		}
		paragraph := ParsedParagraph{HTML: trimmed}
		paragraphs = append(paragraphs, paragraph)
		paragraphCopy := paragraph
		blocks = append(blocks, ParsedBlock{Kind: ParsedBlockKindParagraph, Paragraph: &paragraphCopy})
	}
	if len(blocks) == 0 {
		return Result{}, nil, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeArticleExtractFailed,
			Message:  fmt.Sprintf("extract noticia article for %q", descriptor.Name),
			Source:   descriptor.Name,
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	article := ParsedArticle{
		Dictionary:   "Noticia RAE",
		Lemma:        title,
		CanonicalURL: strings.TrimSpace(document.URL),
		Sections:     []ParsedSection{{Blocks: blocks, Paragraphs: paragraphs}},
	}
	return Result{Articles: []ParsedArticle{article}}, nil, nil
}

func normalizeNoticiaText(raw string) string {
	text := html.UnescapeString(strings.TrimSpace(raw))
	text = reNoticiaTags.ReplaceAllString(text, " ")
	text = strings.ReplaceAll(text, "\u00a0", " ")
	return strings.Join(strings.Fields(text), " ")
}
