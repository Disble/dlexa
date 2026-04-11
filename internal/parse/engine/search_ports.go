package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// SearchParser parses search/discovery fetched documents.
type SearchParser interface {
	ParseSearch(ctx context.Context, input ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error)
}
