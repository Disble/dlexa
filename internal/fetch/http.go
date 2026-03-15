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

	req, err := f.buildRequest(ctx, lookupURL)
	if err != nil {
		return Document{}, model.NewProblemError(model.Problem{
			Code:     model.ProblemCodeDPDFetchFailed,
			Message:  fmt.Sprintf("build DPD request: %v", err),
			Source:   sourceName(request.Source),
			Severity: model.ProblemSeverityError,
		}, err)
	}

	resp, err := f.resolveClient().Do(req)
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

	return f.buildDocument(resp, body, lookupURL), nil
}

// buildRequest creates an HTTP GET request for the given URL and sets default headers.
func (f *DPDFetcher) buildRequest(ctx context.Context, lookupURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupURL, nil)
	if err != nil {
		return nil, err
	}

	if f.UserAgent != "" {
		req.Header.Set("User-Agent", f.UserAgent)
	}

	setDefaultHeader(req, "Accept", "text/html,application/xhtml+xml")
	setDefaultHeader(req, "Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
	setDefaultHeader(req, "Cache-Control", "no-cache")
	setDefaultHeader(req, "Pragma", "no-cache")

	return req, nil
}

// resolveClient returns the configured client or a default one if nil.
func (f *DPDFetcher) resolveClient() Doer {
	if f.Client != nil {
		return f.Client
	}

	return &http.Client{Timeout: 10 * time.Second}
}

// buildDocument assembles a Document from the response, body, and fallback URL.
func (f *DPDFetcher) buildDocument(resp *http.Response, body []byte, fallbackURL string) Document {
	retrievedAt := time.Now().UTC()
	if f.now != nil {
		retrievedAt = f.now()
	}

	finalURL := fallbackURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return Document{
		URL:         finalURL,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		Body:        body,
		RetrievedAt: retrievedAt,
	}
}

// setDefaultHeader sets the header only if it has not already been set.
func setDefaultHeader(req *http.Request, key, value string) {
	if req.Header.Get(key) == "" {
		req.Header.Set(key, value)
	}
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
