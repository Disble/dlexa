package fetch

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type StaticFetcher struct {
	BaseURL     string
	ContentType string
	Template    string
}

func NewStaticFetcher(baseURL string) *StaticFetcher {
	return &StaticFetcher{
		BaseURL:     baseURL,
		ContentType: "text/markdown",
		Template:    "# %s\n\nBootstrap content emitted by the fetch layer.",
	}
}

func (f *StaticFetcher) Fetch(ctx context.Context, request Request) (Document, error) {
	return f.FetchDocument(ctx, request)
}

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
		Body:        []byte(body),
		RetrievedAt: time.Now().UTC(),
	}, nil
}
