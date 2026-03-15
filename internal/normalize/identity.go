package normalize

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// IdentityNormalizer preserves parsed fields without transformation.
type IdentityNormalizer struct{}

// NewIdentityNormalizer returns a new IdentityNormalizer instance.
func NewIdentityNormalizer() *IdentityNormalizer {
	return &IdentityNormalizer{}
}

// Normalize delegates to NormalizeEntries for identity passthrough.
func (n *IdentityNormalizer) Normalize(ctx context.Context, descriptor model.SourceDescriptor, result parse.Result) ([]model.Entry, []model.Warning, error) {
	return n.NormalizeEntries(ctx, descriptor, result)
}

// NormalizeEntries converts parsed articles into entries without semantic transformation.
func (n *IdentityNormalizer) NormalizeEntries(ctx context.Context, descriptor model.SourceDescriptor, result parse.Result) ([]model.Entry, []model.Warning, error) {
	_ = ctx
	normalized := make([]model.Entry, 0, len(result.Articles))
	for _, article := range result.Articles {
		entry := model.Entry{
			ID:       descriptor.Name + ":bootstrap",
			Headword: article.Lemma,
			Summary:  "Parsed bootstrap entry from markdown content.",
			Content:  articleText(article),
			Source:   descriptor.Name,
			URL:      article.CanonicalURL,
			Article: &model.Article{
				Dictionary:   article.Dictionary,
				Edition:      article.Edition,
				Lemma:        article.Lemma,
				CanonicalURL: article.CanonicalURL,
				Sections:     articleSections(article.Sections),
			},
			Metadata: map[string]string{},
		}
		entry.Source = descriptor.Name
		if entry.Metadata == nil {
			entry.Metadata = map[string]string{}
		}
		entry.Metadata["normalized_by"] = "identity"
		normalized = append(normalized, entry)
	}

	warnings := []model.Warning{{
		Code:    "identity_normalizer",
		Message: "normalizer preserves parsed fields until canonical rules exist",
		Source:  descriptor.Name,
	}}

	return normalized, warnings, nil
}

func articleText(article parse.ParsedArticle) string {
	if len(article.Sections) == 0 || len(article.Sections[0].Paragraphs) == 0 {
		return ""
	}

	return article.Sections[0].Paragraphs[0].HTML
}

func articleSections(sections []parse.ParsedSection) []model.Section {
	normalized := make([]model.Section, 0, len(sections))
	for _, section := range sections {
		paragraphs := make([]model.Paragraph, 0, len(section.Paragraphs))
		blocks := make([]model.Block, 0, len(section.Paragraphs))
		for _, paragraph := range section.Paragraphs {
			normalized := model.Paragraph{Markdown: paragraph.HTML}
			paragraphs = append(paragraphs, normalized)
			paragraphCopy := normalized
			blocks = append(blocks, model.Block{Kind: model.ArticleBlockKindParagraph, Paragraph: &paragraphCopy})
		}

		normalized = append(normalized, model.Section{
			Label:      section.Label,
			Title:      section.Title,
			Blocks:     blocks,
			Paragraphs: paragraphs,
			Children:   articleSections(section.Children),
		})
	}

	return normalized
}
