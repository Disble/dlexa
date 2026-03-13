package parse

import (
	"context"
	"strings"

	"github.com/gentleman-programming/dlexa/internal/fetch"
	"github.com/gentleman-programming/dlexa/internal/model"
)

type MarkdownParser struct{}

func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

func (p *MarkdownParser) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (Result, []model.Warning, error) {
	return p.ParseDocument(ctx, descriptor, document)
}

func (p *MarkdownParser) ParseDocument(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (Result, []model.Warning, error) {
	_ = ctx
	body := strings.TrimSpace(string(document.Body))
	headword := descriptor.DisplayName
	if headword == "" {
		headword = descriptor.Name
	}

	warnings := []model.Warning{{
		Code:    "bootstrap_parser",
		Message: "parser is a stub and emits one normalized candidate entry",
		Source:  descriptor.Name,
	}}

	return Result{Articles: []ParsedArticle{{
		Dictionary:   headword,
		Lemma:        headword,
		CanonicalURL: document.URL,
		Sections: []ParsedSection{{
			Label:      "1.",
			Blocks:     []ParsedBlock{{Kind: ParsedBlockKindParagraph, Paragraph: &ParsedParagraph{HTML: body}}},
			Paragraphs: []ParsedParagraph{{HTML: body}},
		}},
	}}}, warnings, nil
}
