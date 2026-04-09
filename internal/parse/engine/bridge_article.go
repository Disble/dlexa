package engine

import (
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// LegacyArticleAdapter adapts the legacy article parser contract into the parser engine port.
type LegacyArticleAdapter struct {
	legacy parse.Parser
}

// AdaptLegacyArticleParser wraps a legacy parser as an engine article parser.
func AdaptLegacyArticleParser(parser parse.Parser) *LegacyArticleAdapter {
	return &LegacyArticleAdapter{legacy: parser}
}

// ParseArticle satisfies the engine article parser port.
func (a *LegacyArticleAdapter) ParseArticle(input ParseInput) (ArticleResult, []model.Warning, error) {
	if a == nil || a.legacy == nil {
		return parse.Result{}, nil, nil
	}
	return a.legacy.Parse(input.Ctx, input.Descriptor, input.Document)
}

// UnderlyingForTesting exposes the wrapped legacy parser.
func (a *LegacyArticleAdapter) UnderlyingForTesting() parse.Parser {
	if a == nil {
		return nil
	}
	return a.legacy
}
