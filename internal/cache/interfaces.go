package cache

import (
	"context"

	"github.com/Disble/dlexa/internal/model"
)

// Store abstracts cache read/write operations for lookup results.
//
// Get returns either a hit, a miss, or a degraded miss when the backing cache
// cannot provide a usable entry. Runtime callers must treat cache errors as
// non-fatal optimization failures and continue with the origin lookup path.
// Set is best-effort: callers may log or inspect write failures, but must not
// fail the user request when fresh data was produced successfully.
type Store interface {
	Get(ctx context.Context, key string) (model.LookupResult, bool, error)
	Set(ctx context.Context, key string, result model.LookupResult) error
}
