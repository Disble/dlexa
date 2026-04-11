package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	legacyparse "github.com/Disble/dlexa/internal/parse"
)

// DPDArticleParser is the engine-native wrapper for the DPD article parser.
type DPDArticleParser struct {
	legacy *legacyparse.DPDArticleParser
}

// NewDPDArticleParser creates an engine-native wrapper for DPD article parsing.
func NewDPDArticleParser() *DPDArticleParser {
	return &DPDArticleParser{legacy: legacyparse.NewDPDArticleParser()}
}

// ParseArticle delegates to the legacy parser implementation without behavior changes.
func (p *DPDArticleParser) ParseArticle(ctx context.Context, input ParseInput) (ArticleResult, []model.Warning, error) {
	return p.legacy.Parse(ctx, input.Descriptor, input.Document)
}

// UnderlyingForTesting exposes the legacy parser implementation.
func (p *DPDArticleParser) UnderlyingForTesting() *legacyparse.DPDArticleParser {
	if p == nil {
		return nil
	}
	return p.legacy
}
