// Package query implements the lookup service layer for dlexa.
package query

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
)

// Looker defines the contract for performing dictionary lookups.
type Looker interface {
	Lookup(ctx context.Context, request model.LookupRequest) (model.LookupResult, error)
}
