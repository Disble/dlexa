package fetch

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// StaticFetcher returns synthetic documents for bootstrap and testing.
type StaticFetcher struct {
	BaseURL     string
	ContentType string
	Template    string
}

// NewStaticFetcher creates a StaticFetcher with the given base URL.
func NewStaticFetcher(baseURL string) *StaticFetcher {
	return &StaticFetcher{
		BaseURL:     baseURL,
		ContentType: "text/markdown",
		Template:    "# %s\n\nBootstrap content emitted by the fetch layer.",
	}
}

// Fetch delegates to FetchDocument to satisfy the Fetcher interface.
func (f *StaticFetcher) Fetch(ctx context.Context, request Request) (Document, error) {
	return f.FetchDocument(ctx, request)
}

// FetchDocument generates a synthetic document from the request query.
func (f *StaticFetcher) FetchDocument(ctx context.Context, request Request) (Document, error) {
	_ = ctx
	query := strings.TrimSpace(request.Query)
	if query == "" {
		query = "unknown"
	}

	body := fmt.Sprintf(f.Template, query)

	return Document{
		URL:         strings.TrimRight(f.BaseURL, "/") + "/" + query,
		ContentType: f.ContentType,
		StatusCode:  200,
		Body:        []byte(body),
		RetrievedAt: time.Now().UTC(),
	}, nil
}
