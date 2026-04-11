package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	legacyparse "github.com/Disble/dlexa/internal/parse"
)

// NoticiaArticleParser is the engine-native wrapper for the noticia parser.
type NoticiaArticleParser struct {
	legacy *legacyparse.NoticiaParser
}

// NewNoticiaArticleParser creates an engine-native wrapper for noticia article parsing.
func NewNoticiaArticleParser() *NoticiaArticleParser {
	return &NoticiaArticleParser{legacy: legacyparse.NewNoticiaParser()}
}

// ParseArticle delegates to the legacy parser implementation without behavior changes.
func (p *NoticiaArticleParser) ParseArticle(ctx context.Context, input ParseInput) (ArticleResult, []model.Warning, error) {
	return p.legacy.Parse(ctx, input.Descriptor, input.Document)
}
