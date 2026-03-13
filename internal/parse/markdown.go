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

func (p *MarkdownParser) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]model.Entry, []model.Warning, error) {
	return p.ParseDocument(ctx, descriptor, document)
}

func (p *MarkdownParser) ParseDocument(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]model.Entry, []model.Warning, error) {
	_ = ctx
	body := strings.TrimSpace(string(document.Body))
	headword := descriptor.DisplayName
	if headword == "" {
		headword = descriptor.Name
	}

	entry := model.Entry{
		ID:       descriptor.Name + ":bootstrap",
		Headword: headword,
		Summary:  "Parsed bootstrap entry from markdown content.",
		Content:  body,
		Source:   descriptor.Name,
		URL:      document.URL,
		Metadata: map[string]string{
			"content_type": document.ContentType,
		},
	}

	warnings := []model.Warning{{
		Code:    "bootstrap_parser",
		Message: "parser is a stub and emits one normalized candidate entry",
		Source:  descriptor.Name,
	}}

	return []model.Entry{entry}, warnings, nil
}
