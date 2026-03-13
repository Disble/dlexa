package parse

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/fetch"
	"github.com/gentleman-programming/dlexa/internal/model"
)

type Parser interface {
	Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]model.Entry, []model.Warning, error)
}
