package query

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type Service interface {
	Lookup(ctx context.Context, request model.LookupRequest) (model.LookupResult, error)
}
