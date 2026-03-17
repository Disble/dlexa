// Package normalize transforms parsed data into canonical domain entries.
package normalize

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// Normalizer converts parse results into normalized model entries.
type Normalizer interface {
	Normalize(ctx context.Context, descriptor model.SourceDescriptor, result parse.Result) (Result, error)
}

// Result holds normalized entries, structured miss data, and non-fatal warnings.
type Result struct {
	Entries  []model.Entry
	Miss     *model.LookupMiss
	Warnings []model.Warning
}
