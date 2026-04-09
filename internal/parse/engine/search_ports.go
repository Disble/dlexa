package engine

import (
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// SearchParser parses search/discovery fetched documents.
type SearchParser interface {
	ParseSearch(input ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error)
}
