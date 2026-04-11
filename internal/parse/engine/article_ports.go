package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// ArticleResult is the article-family parser result.
type ArticleResult = parse.Result

// ArticleParser parses article-like fetched documents.
type ArticleParser interface {
	ParseArticle(ctx context.Context, input ParseInput) (ArticleResult, []model.Warning, error)
}
