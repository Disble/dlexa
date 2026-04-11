package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	legacyparse "github.com/Disble/dlexa/internal/parse"
)

// DudaLinguisticaArticleParser is the engine-native wrapper for the Duda lingüística parser.
type DudaLinguisticaArticleParser struct {
	legacy *legacyparse.DudaLinguisticaParser
}

// NewDudaLinguisticaArticleParser creates an engine-native wrapper for duda-linguistica article parsing.
func NewDudaLinguisticaArticleParser() *DudaLinguisticaArticleParser {
	return &DudaLinguisticaArticleParser{legacy: legacyparse.NewDudaLinguisticaParser()}
}

// ParseArticle delegates to the legacy parser implementation without behavior changes.
func (p *DudaLinguisticaArticleParser) ParseArticle(ctx context.Context, input ParseInput) (ArticleResult, []model.Warning, error) {
	return p.legacy.Parse(ctx, input.Descriptor, input.Document)
}
