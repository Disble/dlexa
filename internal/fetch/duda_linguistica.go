package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

	req, err := buildDPDRequest(ctx, http.MethodGet, articleURL, f.UserAgent, "text/html,application/xhtml+xml")
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("build duda-linguistica request: %v", err), request.Source, err)
	}

	resp, err := resolveClient(f.Client).Do(req)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("fetch duda-linguistica document: %v", err), request.Source, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("read duda-linguistica response body: %v", err), request.Source, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return Document{}, newFetchProblem(model.ProblemCodeArticleNotFound, fmt.Sprintf("article not found for %q", normalizeQuery(request.Query)), request.Source, nil)
	}
	if isChallengeBody(body) {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, "duda-linguistica request was challenged by upstream; browser-like profile still rejected", request.Source, nil)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("duda-linguistica request failed with status %d", resp.StatusCode), request.Source, nil)
	}

	return buildDocument(f.now, resp, body, articleURL), nil
}

func (f *DudaLinguisticaFetcher) lookupURL(rawQuery string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("duda-linguistica base URL is empty")
	}

	slug := normalizeQuery(rawQuery)
	if slug == "" {
		return "", fmt.Errorf("article slug is empty")
	}

	base, err := url.Parse(baseURL + "/")
	if err != nil {
		return "", fmt.Errorf("parse duda-linguistica base URL: %w", err)
	}
	relative, err := url.Parse("../duda-linguistica/" + url.PathEscape(slug))
	if err != nil {
		return "", fmt.Errorf("encode duda-linguistica slug: %w", err)
	}
	return base.ResolveReference(relative).String(), nil
}
