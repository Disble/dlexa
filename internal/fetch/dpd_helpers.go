package fetch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Disble/dlexa/internal/model"
)

const challengeBodySnippetLimit = 1024

func buildDPDRequest(ctx context.Context, method, targetURL, userAgent, accept string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, targetURL, nil)
	if err != nil {
		return nil, err
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	setDefaultHeader(req, "Accept", accept)
	setDefaultHeader(req, "Accept-Language", "es-ES,es;q=0.9,en;q=0.8")
	setDefaultHeader(req, "Cache-Control", "no-cache")
	setDefaultHeader(req, "Pragma", "no-cache")
	return req, nil
}

func resolveClient(client Doer) Doer {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func buildDocument(nowFn func() time.Time, resp *http.Response, body []byte, lookupURL string) Document {
	retrievedAt := time.Now().UTC()
	if nowFn != nil {
		retrievedAt = nowFn()
	}
	// resp.Request.URL is the final URL after the Go HTTP client followed any
	// server-side redirects (e.g. /dpd/solo → /dpd/tilde). We surface both
	// URLs so the parser can derive the correct user query term from the original
	// URL and warn the user that a redirect occurred.
	finalURL := lookupURL
	if resp.Request != nil && resp.Request.URL != nil {
		if u := resp.Request.URL.String(); u != "" {
			finalURL = u
		}
	}
	doc := Document{
		URL:         finalURL,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		Body:        body,
		RetrievedAt: retrievedAt,
	}
	if finalURL != lookupURL {
		doc.RedirectedFrom = lookupURL
	}
	return doc
}

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

func normalizeQuery(raw string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
}

func newFetchProblem(code, message string, descriptor model.SourceDescriptor, err error) error {
	return model.NewProblemError(model.Problem{Code: code, Message: message, Source: sourceName(descriptor), Severity: model.ProblemSeverityError}, err)
}

func sourceName(descriptor model.SourceDescriptor) string {
	if strings.TrimSpace(descriptor.Name) != "" {
		return descriptor.Name
	}
	return "dpd"
}

func emptyQueryError(kind string) error {
	return fmt.Errorf("dpd %s term is empty", kind)
}
