package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

// DPDSearchFetcher retrieves entry-search payloads from the DPD keys endpoint.
type DPDSearchFetcher struct {
	BaseURL   string
	UserAgent string
	Client    Doer
	now       func() time.Time
}

// NewDPDSearchFetcher creates a DPDSearchFetcher with the given base URL, timeout, and user agent.
func NewDPDSearchFetcher(baseURL string, timeout time.Duration, userAgent string) *DPDSearchFetcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &DPDSearchFetcher{
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		UserAgent: strings.TrimSpace(userAgent),
		Client:    &http.Client{Timeout: timeout},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Fetch retrieves a DPD keys payload for the given request query.
func (f *DPDSearchFetcher) Fetch(ctx context.Context, request Request) (Document, error) {
	if f == nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, "dpd search fetcher is not configured", request.Source, nil)
	}

	searchURL, err := f.searchURL(request.Query)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, err.Error(), request.Source, err)
	}

	req, err := buildDPDRequest(ctx, http.MethodGet, searchURL, f.UserAgent, "application/json,text/plain,*/*")
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, fmt.Sprintf("build DPD entry search request: %v", err), request.Source, err)
	}

	resp, err := resolveClient(f.Client).Do(req)
	if err != nil {
		var cooldownErr *RateLimitCooldownError
		if errors.As(err, &cooldownErr) {
			return Document{}, newFetchProblem(model.ProblemCodeDPDSearchRateLimited, fmt.Sprintf("DPD entry search temporarily rate-limited: dpd transport cooling down for %s after upstream rate limiting", cooldownErr.Remaining.Round(time.Second)), request.Source, err)
		}
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, fmt.Sprintf("fetch DPD entry search payload: %v", err), request.Source, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, fmt.Sprintf("read DPD entry search response body: %v", err), request.Source, err)
	}

	if isChallengeBody(body) {
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, "DPD entry search was challenged by upstream; browser-like profile still rejected", request.Source, nil)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if resp.StatusCode == http.StatusTooManyRequests {
			return Document{}, newFetchProblem(model.ProblemCodeDPDSearchRateLimited, "DPD entry search was rate-limited by upstream (status 429)", request.Source, nil)
		}
		return Document{}, newFetchProblem(model.ProblemCodeDPDSearchFetchFailed, fmt.Sprintf("DPD entry search request failed with status %d", resp.StatusCode), request.Source, nil)
	}

	return buildDocument(f.now, resp, body, searchURL), nil
}

func (f *DPDSearchFetcher) searchURL(rawQuery string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("dpd base URL is empty")
	}

	term := normalizeQuery(rawQuery)
	if term == "" {
		return "", emptyQueryError("entry search")
	}

	base, err := url.Parse(baseURL + "/srv/keys")
	if err != nil {
		return "", fmt.Errorf("parse DPD base URL: %w", err)
	}
	query := base.Query()
	query.Set("q", term)
	base.RawQuery = query.Encode()

	return base.String(), nil
}
