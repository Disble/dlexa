package fetch

import (
	"context"
	"time"

	"github.com/gentleman-programming/dlexa/internal/model"
)

// Request holds the query and source metadata for a fetch operation.
type Request struct {
	Query  string
	Source model.SourceDescriptor
}

// Document represents a fetched document with its metadata and body.
type Document struct {
	URL         string
	ContentType string
	StatusCode  int
	Body        []byte
	RetrievedAt time.Time
}

// Fetcher retrieves documents from a dictionary source.
type Fetcher interface {
	Fetch(ctx context.Context, request Request) (Document, error)
}
