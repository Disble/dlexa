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

	req, err := buildDPDRequest(ctx, http.MethodGet, articleURL, f.UserAgent, "text/html,application/xhtml+xml")
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("build noticia request: %v", err), request.Source, err)
	}

	resp, err := resolveClient(f.Client).Do(req)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("fetch noticia document: %v", err), request.Source, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("read noticia response body: %v", err), request.Source, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return Document{}, newFetchProblem(model.ProblemCodeArticleNotFound, fmt.Sprintf("article not found for %q", normalizeQuery(request.Query)), request.Source, nil)
	}
	if isChallengeBody(body) {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, "noticia request was challenged by upstream; browser-like profile still rejected", request.Source, nil)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("noticia request failed with status %d", resp.StatusCode), request.Source, nil)
	}

	return buildDocument(f.now, resp, body, articleURL), nil
}

func (f *NoticiaFetcher) lookupURL(rawQuery string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("noticia base URL is empty")
	}

	slug := normalizeQuery(rawQuery)
	if slug == "" {
		return "", fmt.Errorf("article slug is empty")
	}

	base, err := url.Parse(baseURL + "/")
	if err != nil {
		return "", fmt.Errorf("parse noticia base URL: %w", err)
	}
	relative, err := url.Parse("../noticia/" + url.PathEscape(slug))
	if err != nil {
		return "", fmt.Errorf("encode noticia slug: %w", err)
	}
	return base.ResolveReference(relative).String(), nil
}
