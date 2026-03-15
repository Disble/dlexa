package cache

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
)

// Store abstracts cache read/write operations for lookup results.
type Store interface {
	Get(ctx context.Context, key string) (model.LookupResult, bool, error)
	Set(ctx context.Context, key string, result model.LookupResult) error
}
