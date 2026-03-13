package cache

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type Store interface {
	Get(ctx context.Context, key string) (model.LookupResult, bool, error)
	Set(ctx context.Context, key string, result model.LookupResult) error
}
