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

// Doer abstracts the HTTP transport for testability.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// DPDFetcher retrieves dictionary entries from the DPD website.
type DPDFetcher struct {
	BaseURL   string
	UserAgent string
	Client    Doer
	now       func() time.Time
}

// NewDPDFetcher creates a DPDFetcher with the given base URL, timeout, and user agent.
func NewDPDFetcher(baseURL string, timeout time.Duration, userAgent string) *DPDFetcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &DPDFetcher{
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		UserAgent: strings.TrimSpace(userAgent),
		Client: &http.Client{
			Timeout: timeout,
		},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Fetch retrieves a DPD document for the given request query.
func (f *DPDFetcher) Fetch(ctx context.Context, request Request) (Document, error) {
	if f == nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDFetchFailed, "dpd fetcher is not configured", request.Source, nil)
	}

	lookupURL, err := f.lookupURL(request.Query)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDFetchFailed, err.Error(), request.Source, err)
	}

	req, err := buildDPDRequest(ctx, http.MethodGet, lookupURL, f.UserAgent, "text/html,application/xhtml+xml")
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDFetchFailed, fmt.Sprintf("build DPD request: %v", err), request.Source, err)
	}

	resp, err := resolveClient(f.Client).Do(req)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDFetchFailed, fmt.Sprintf("fetch DPD document: %v", err), request.Source, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Document{}, newFetchProblem(model.ProblemCodeDPDFetchFailed, fmt.Sprintf("read DPD response body: %v", err), request.Source, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return Document{}, newFetchProblem(model.ProblemCodeDPDNotFound, fmt.Sprintf("DPD entry not found for %q", normalizeQuery(request.Query)), request.Source, nil)
	}

	if isChallengeBody(body) {
		return Document{}, newFetchProblem(model.ProblemCodeDPDFetchFailed, "DPD request was challenged by upstream; browser-like profile still rejected", request.Source, nil)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Document{}, newFetchProblem(model.ProblemCodeDPDFetchFailed, fmt.Sprintf("DPD request failed with status %d", resp.StatusCode), request.Source, nil)
	}

	return buildDocument(f.now, resp, body, lookupURL), nil
}

func (f *DPDFetcher) lookupURL(rawQuery string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("dpd base URL is empty")
	}

	term := normalizeQuery(rawQuery)
	if term == "" {
		return "", emptyQueryError("lookup")
	}

	base, err := url.Parse(baseURL + "/")
	if err != nil {
		return "", fmt.Errorf("parse DPD base URL: %w", err)
	}

	relative, err := url.Parse(url.PathEscape(term))
	if err != nil {
		return "", fmt.Errorf("encode DPD lookup term: %w", err)
	}

	return base.ResolveReference(relative).String(), nil
}
