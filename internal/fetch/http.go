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

const challengeBodySnippetLimit = 1024

// HTTPClient abstracts the HTTP transport for testability.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DPDFetcher retrieves dictionary entries from the DPD website.
type DPDFetcher struct {
	BaseURL   string
	UserAgent string
	Client    HTTPClient
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
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDFetchFailed,
			Message:  "dpd fetcher is not configured",
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	lookupURL, err := f.lookupURL(request.Query)
	if err != nil {
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDFetchFailed,
			Message:  err.Error(),
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, err)
	}

	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupURL, nil)
	if err != nil {
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDFetchFailed,
			Message:  fmt.Sprintf("build DPD request: %v", err),
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, err)
	}

	if f.UserAgent != "" {
		req.Header.Set("User-Agent", f.UserAgent)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
	}
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
	}
	if req.Header.Get("Cache-Control") == "" {
		req.Header.Set("Cache-Control", "no-cache")
	}
	if req.Header.Get("Pragma") == "" {
		req.Header.Set("Pragma", "no-cache")
	}

	resp, err := client.Do(req)
	if err != nil {
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDFetchFailed,
			Message:  fmt.Sprintf("fetch DPD document: %v", err),
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDFetchFailed,
			Message:  fmt.Sprintf("read DPD response body: %v", err),
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDNotFound,
			Message:  fmt.Sprintf("DPD entry not found for %q", normalizeQuery(request.Query)),
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	if isChallengeBody(body) {
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDFetchFailed,
			Message:  "DPD request was challenged by upstream; browser-like profile still rejected",
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDFetchFailed,
			Message:  fmt.Sprintf("DPD request failed with status %d", resp.StatusCode),
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, nil)
	}

	retrievedAt := time.Now().UTC()
	if f.now != nil {
		retrievedAt = f.now()
	}

	finalURL := lookupURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return Document{
		URL:         finalURL,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		Body:        body,
		RetrievedAt: retrievedAt,
	}, nil
}

func isChallengeBody(body []byte) bool {
	limit := len(body)
	if limit > challengeBodySnippetLimit {
		limit = challengeBodySnippetLimit
	}
	snippet := strings.ToLower(string(body[:limit]))
	return strings.Contains(snippet, "cloudflare") && strings.Contains(snippet, "challenge")
}

func (f *DPDFetcher) lookupURL(rawQuery string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("dpd base URL is empty")
	}

	term := normalizeQuery(rawQuery)
	if term == "" {
		return "", fmt.Errorf("dpd lookup term is empty")
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

func normalizeQuery(raw string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
}

func sourceName(descriptor model.SourceDescriptor) string {
	if strings.TrimSpace(descriptor.Name) != "" {
		return descriptor.Name
	}

	return "dpd"
}
