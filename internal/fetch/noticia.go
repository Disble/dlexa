package fetch

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

// NoticiaFetcher retrieves a concrete noticia page.
type NoticiaFetcher struct {
	BaseURL   string
	UserAgent string
	Client    Doer
	now       func() time.Time
}

// NewNoticiaFetcher creates an article fetcher for the noticia surface.
func NewNoticiaFetcher(baseURL string, timeout time.Duration, userAgent string) *NoticiaFetcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &NoticiaFetcher{
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		UserAgent: strings.TrimSpace(userAgent),
		Client:    &http.Client{Timeout: timeout},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Fetch retrieves one noticia document by slug.
func (f *NoticiaFetcher) Fetch(ctx context.Context, request Request) (Document, error) {
	if f == nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, "noticia fetcher is not configured", request.Source, nil)
	}

	articleURL, err := f.lookupURL(request.Query)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, err.Error(), request.Source, err)
	}

	return fetchArticleDocument(ctx, f.Client, f.now, f.UserAgent, articleURL, moduleNameNoticia, request)
}

func (f *NoticiaFetcher) lookupURL(rawQuery string) (string, error) {
	return buildArticleLookupURL(f.BaseURL, moduleNameNoticia, rawQuery)
}

const moduleNameNoticia = "noticia"
