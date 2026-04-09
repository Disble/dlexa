package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// LegacySearchParser is the legacy search parser contract mirrored here to avoid package cycles.
type LegacySearchParser interface {
	Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]parse.ParsedSearchRecord, []model.Warning, error)
}

// LegacySearchAdapter adapts the legacy search parser contract into the parser engine port.
type LegacySearchAdapter struct {
	legacy LegacySearchParser
}

// AdaptLegacySearchParser wraps a legacy search parser as an engine search parser.
func AdaptLegacySearchParser(parser LegacySearchParser) *LegacySearchAdapter {
	return &LegacySearchAdapter{legacy: parser}
}

// ParseSearch satisfies the engine search parser port.
func (a *LegacySearchAdapter) ParseSearch(input ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error) {
	if a == nil || a.legacy == nil {
		return nil, nil, nil
	}
	return a.legacy.Parse(input.Ctx, input.Descriptor, input.Document)
}

// UnderlyingForTesting exposes the wrapped legacy parser.
func (a *LegacySearchAdapter) UnderlyingForTesting() LegacySearchParser {
	if a == nil {
		return nil
	}
	return a.legacy
}
