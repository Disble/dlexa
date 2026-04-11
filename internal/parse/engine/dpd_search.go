package engine

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	legacyparse "github.com/Disble/dlexa/internal/parse"
)

// DPDSearchParser is the engine-native wrapper for the DPD search parser.
type DPDSearchParser struct {
	legacy *legacyparse.DPDSearchParser
}

// NewDPDSearchParser creates an engine-native wrapper for DPD search parsing.
func NewDPDSearchParser() *DPDSearchParser {
	return &DPDSearchParser{legacy: legacyparse.NewDPDSearchParser()}
}

// ParseSearch delegates to the legacy parser implementation without behavior changes.
func (p *DPDSearchParser) ParseSearch(ctx context.Context, input ParseInput) ([]legacyparse.ParsedSearchRecord, []model.Warning, error) {
	return p.legacy.Parse(ctx, input.Descriptor, input.Document)
}

// UnderlyingForTesting exposes the legacy parser implementation.
func (p *DPDSearchParser) UnderlyingForTesting() *legacyparse.DPDSearchParser {
	if p == nil {
		return nil
	}
	return p.legacy
}
