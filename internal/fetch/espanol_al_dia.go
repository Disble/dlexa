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

	req, err := buildDPDRequest(ctx, http.MethodGet, articleURL, f.UserAgent, "text/html,application/xhtml+xml")
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("build espanol-al-dia request: %v", err), request.Source, err)
	}

	resp, err := resolveClient(f.Client).Do(req)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("fetch espanol-al-dia document: %v", err), request.Source, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("read espanol-al-dia response body: %v", err), request.Source, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return Document{}, newFetchProblem(model.ProblemCodeArticleNotFound, fmt.Sprintf("article not found for %q", normalizeQuery(request.Query)), request.Source, nil)
	}
	if isChallengeBody(body) {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, "espanol-al-dia request was challenged by upstream; browser-like profile still rejected", request.Source, nil)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Document{}, newFetchProblem(model.ProblemCodeArticleFetchFailed, fmt.Sprintf("espanol-al-dia request failed with status %d", resp.StatusCode), request.Source, nil)
	}

	return buildDocument(f.now, resp, body, articleURL), nil
}

func (f *EspanolAlDiaFetcher) lookupURL(rawQuery string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("espanol-al-dia base URL is empty")
	}

	slug := normalizeQuery(rawQuery)
	if slug == "" {
		return "", fmt.Errorf("article slug is empty")
	}

	base, err := url.Parse(baseURL + "/")
	if err != nil {
		return "", fmt.Errorf("parse espanol-al-dia base URL: %w", err)
	}
	relative, err := url.Parse("../espanol-al-dia/" + url.PathEscape(slug))
	if err != nil {
		return "", fmt.Errorf("encode espanol-al-dia slug: %w", err)
	}
	return base.ResolveReference(relative).String(), nil
}
