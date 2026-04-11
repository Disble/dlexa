package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	legacyparse "github.com/Disble/dlexa/internal/parse"
)

// EspanolAlDiaArticleParser is the engine-native wrapper for the Español al día parser.
type EspanolAlDiaArticleParser struct {
	legacy *legacyparse.EspanolAlDiaParser
}

// NewEspanolAlDiaArticleParser creates an engine-native wrapper for español-al-día article parsing.
func NewEspanolAlDiaArticleParser() *EspanolAlDiaArticleParser {
	return &EspanolAlDiaArticleParser{legacy: legacyparse.NewEspanolAlDiaParser()}
}

// ParseArticle delegates to the legacy parser implementation without behavior changes.
func (p *EspanolAlDiaArticleParser) ParseArticle(ctx context.Context, input ParseInput) (ArticleResult, []model.Warning, error) {
	return p.legacy.Parse(ctx, input.Descriptor, input.Document)
}
