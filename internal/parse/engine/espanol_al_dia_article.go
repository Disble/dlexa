package engine

import (
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
func (p *EspanolAlDiaArticleParser) ParseArticle(input ParseInput) (ArticleResult, []model.Warning, error) {
	return p.legacy.Parse(input.Ctx, input.Descriptor, input.Document)
}
