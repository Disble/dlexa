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

// LiveSearchFetcher retrieves the general RAE search results page.
type LiveSearchFetcher struct {
	BaseURL   string
	UserAgent string
	Client    Doer
	now       func() time.Time
}

// NewLiveSearchFetcher creates a LiveSearchFetcher with bounded HTTP behavior.
func NewLiveSearchFetcher(baseURL string, timeout time.Duration, userAgent string) *LiveSearchFetcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &LiveSearchFetcher{
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		UserAgent: strings.TrimSpace(userAgent),
		Client:    &http.Client{Timeout: timeout},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Fetch retrieves the live RAE search results page for the given query.
func (f *LiveSearchFetcher) Fetch(ctx context.Context, request Request) (Document, error) {
	if f == nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, "live search fetcher is not configured", request.Source, nil)
	}

	searchURL, err := f.searchURL(request.Query)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, err.Error(), request.Source, err)
	}

	req, err := buildDPDRequest(ctx, http.MethodGet, searchURL, f.UserAgent, "text/html,application/xhtml+xml")
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, fmt.Sprintf("build live search request: %v", err), request.Source, err)
	}

	resp, err := resolveClient(f.Client).Do(req)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, fmt.Sprintf("fetch live search page: %v", err), request.Source, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, fmt.Sprintf("read live search response body: %v", err), request.Source, err)
	}

	if isChallengeBody(body) {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, "live search was challenged by upstream; browser-like profile still rejected", request.Source, nil)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, fmt.Sprintf("live search request failed with status %d", resp.StatusCode), request.Source, nil)
	}

	return buildDocument(f.now, resp, body, searchURL), nil
}

func (f *LiveSearchFetcher) searchURL(rawQuery string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("search base URL is empty")
	}

	term := normalizeQuery(rawQuery)
	if term == "" {
		return "", fmt.Errorf("search term is empty")
	}

	base, err := url.Parse(baseURL + "/")
	if err != nil {
		return "", fmt.Errorf("parse search base URL: %w", err)
	}
	relative, err := url.Parse("../search/node")
	if err != nil {
		return "", fmt.Errorf("resolve live search endpoint: %w", err)
	}
	endpoint := base.ResolveReference(relative)
	query := endpoint.Query()
	query.Set("keys", term)
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}
