package fetch

import (
	"context"
	"time"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type Request struct {
	Query  string
	Source model.SourceDescriptor
}

type Document struct {
	URL         string
	ContentType string
	Body        []byte
	RetrievedAt time.Time
}

type Fetcher interface {
	Fetch(ctx context.Context, request Request) (Document, error)
}
