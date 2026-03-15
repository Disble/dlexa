// Package query implements the lookup service layer for dlexa.
package query

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/model"
)

// Service defines the contract for performing dictionary lookups.
type Service interface {
	Lookup(ctx context.Context, request model.LookupRequest) (model.LookupResult, error)
}
