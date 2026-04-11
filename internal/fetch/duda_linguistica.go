package fetch

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

// DudaLinguisticaFetcher retrieves a concrete Duda lingüística article page.
type DudaLinguisticaFetcher struct {
	BaseURL   string
	UserAgent string
	Client    Doer
	now       func() time.Time
}

// NewDudaLinguisticaFetcher creates an article fetcher for the duda-linguistica surface.
func NewDudaLinguisticaFetcher(baseURL string, timeout time.Duration, userAgent string) *DudaLinguisticaFetcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &DudaLinguisticaFetcher{
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		UserAgent: strings.TrimSpace(userAgent),
		Client:    &http.Client{Timeout: timeout},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Fetch retrieves one duda-linguistica article document by slug.
func (f *DudaLinguisticaFetcher) Fetch(ctx context.Context, request Request) (Document, error) {
	if f == nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, "duda-linguistica fetcher is not configured", request.Source, nil)
	}

	articleURL, err := f.lookupURL(request.Query)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, err.Error(), request.Source, err)
	}

	return fetchArticleDocument(ctx, f.Client, f.now, f.UserAgent, articleURL, moduleNameDudaLinguistica, request)
}

func (f *DudaLinguisticaFetcher) lookupURL(rawQuery string) (string, error) {
	return buildArticleLookupURL(f.BaseURL, moduleNameDudaLinguistica, rawQuery)
}

const moduleNameDudaLinguistica = "duda-linguistica"
