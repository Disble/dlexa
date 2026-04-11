package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	legacyparse "github.com/Disble/dlexa/internal/parse"
)

// LiveSearchParser is the engine-native wrapper for the general search parser.
type LiveSearchParser struct {
	legacy *legacyparse.LiveSearchParser
}

// NewLiveSearchParser creates an engine-native wrapper for live search parsing.
func NewLiveSearchParser() *LiveSearchParser {
	return &LiveSearchParser{legacy: legacyparse.NewLiveSearchParser()}
}

// ParseSearch delegates to the legacy parser implementation without behavior changes.
func (p *LiveSearchParser) ParseSearch(ctx context.Context, input ParseInput) ([]legacyparse.ParsedSearchRecord, []model.Warning, error) {
	return p.legacy.Parse(ctx, input.Descriptor, input.Document)
}

// UnderlyingForTesting exposes the legacy parser implementation.
func (p *LiveSearchParser) UnderlyingForTesting() *legacyparse.LiveSearchParser {
	if p == nil {
		return nil
	}
	return p.legacy
}
