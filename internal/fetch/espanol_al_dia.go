package fetch

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

// EspanolAlDiaFetcher retrieves a concrete Español al día article page.
type EspanolAlDiaFetcher struct {
	BaseURL   string
	UserAgent string
	Client    Doer
	now       func() time.Time
}

// NewEspanolAlDiaFetcher creates an article fetcher for the español-al-día surface.
func NewEspanolAlDiaFetcher(baseURL string, timeout time.Duration, userAgent string) *EspanolAlDiaFetcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &EspanolAlDiaFetcher{
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		UserAgent: strings.TrimSpace(userAgent),
		Client:    &http.Client{Timeout: timeout},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Fetch retrieves one español-al-día article document by slug.
func (f *EspanolAlDiaFetcher) Fetch(ctx context.Context, request Request) (Document, error) {
	if f == nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, "espanol-al-dia fetcher is not configured", request.Source, nil)
	}

	articleURL, err := f.lookupURL(request.Query)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, err.Error(), request.Source, err)
	}

	return fetchArticleDocument(ctx, f.Client, f.now, f.UserAgent, articleURL, moduleNameEspanolAlDia, request)
}

func (f *EspanolAlDiaFetcher) lookupURL(rawQuery string) (string, error) {
	return buildArticleLookupURL(f.BaseURL, moduleNameEspanolAlDia, rawQuery)
}

const moduleNameEspanolAlDia = "espanol-al-dia"
