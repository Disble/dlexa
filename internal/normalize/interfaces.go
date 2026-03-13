package normalize

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type Normalizer interface {
	Normalize(ctx context.Context, descriptor model.SourceDescriptor, entries []model.Entry) ([]model.Entry, []model.Warning, error)
}
