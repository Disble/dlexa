package fetch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gentleman-programming/dlexa/internal/model"
)

func TestDPDFetcherClassifiesTransportOutcomesAndCapturesDocuments(t *testing.T) {
	fixedNow := time.Date(2026, time.March, 13, 16, 30, 0, 0, time.UTC)
	request := Request{Query: "  bien compuesto  ", Source: model.SourceDescriptor{Name: "dpd"}}

	tests := []struct {
		name           string
		client         HTTPClient
		wantDocument   Document
		wantProblem    *model.Problem
		wantRequestURL string
		wantUserAgent  string
		wantAcceptLang string
	}{
		{
			name: "successful html document capture",
			client: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got := req.URL.String(); got != "https://example.invalid/dpd/bien%20compuesto" {
					return nil, errors.New("unexpected request URL: " + got)
				}
				if got := req.Header.Get("User-Agent"); got != "dlexa-test" {
					return nil, errors.New("unexpected user agent: " + got)
				}
				if got := req.Header.Get("Accept-Language"); got != "es-ES,es;q=0.9,en;q=0.8" {
					return nil, errors.New("unexpected accept-language: " + got)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/html; charset=utf-8"},
					},
					Body:    io.NopCloser(strings.NewReader("<html>ok</html>")),
					Request: req,
				}, nil
			}),
			wantDocument: Document{
				URL:         "https://example.invalid/dpd/bien%20compuesto",
				ContentType: "text/html; charset=utf-8",
				StatusCode:  http.StatusOK,
				Body:        []byte("<html>ok</html>"),
				RetrievedAt: fixedNow,
			},
			wantRequestURL: "https://example.invalid/dpd/bien%20compuesto",
			wantUserAgent:  "dlexa-test",
			wantAcceptLang: "es-ES,es;q=0.9,en;q=0.8",
		},
		{
			name: "timeout failure becomes fetch problem",
			client: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return nil, context.DeadlineExceeded
			}),
			wantProblem: &model.Problem{
				Code:     model.ProblemCodeDPDFetchFailed,
				Message:  "fetch DPD document: context deadline exceeded",
				Source:   "dpd",
				Severity: model.ProblemSeverityError,
			},
			wantRequestURL: "https://example.invalid/dpd/bien%20compuesto",
			wantUserAgent:  "dlexa-test",
			wantAcceptLang: "es-ES,es;q=0.9,en;q=0.8",
		},
		{
			name: "network failure becomes fetch problem",
			client: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return nil, errors.New("dial tcp connection refused")
			}),
			wantProblem: &model.Problem{
				Code:     model.ProblemCodeDPDFetchFailed,
				Message:  "fetch DPD document: dial tcp connection refused",
				Source:   "dpd",
				Severity: model.ProblemSeverityError,
			},
			wantRequestURL: "https://example.invalid/dpd/bien%20compuesto",
			wantUserAgent:  "dlexa-test",
			wantAcceptLang: "es-ES,es;q=0.9,en;q=0.8",
		},
		{
			name: "404 becomes not found problem",
			client: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("missing")),
					Request:    req,
				}, nil
			}),
			wantProblem: &model.Problem{
				Code:     model.ProblemCodeDPDNotFound,
				Message:  "DPD entry not found for \"bien compuesto\"",
				Source:   "dpd",
				Severity: model.ProblemSeverityError,
			},
			wantRequestURL: "https://example.invalid/dpd/bien%20compuesto",
			wantUserAgent:  "dlexa-test",
			wantAcceptLang: "es-ES,es;q=0.9,en;q=0.8",
		},
		{
			name: "challenge page becomes fetch problem",
			client: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(strings.NewReader("<html><title>Just a moment...</title><div>Cloudflare challenge</div></html>")),
					Request:    req,
				}, nil
			}),
			wantProblem: &model.Problem{
				Code:     model.ProblemCodeDPDFetchFailed,
				Message:  "DPD request was challenged by upstream; browser-like profile still rejected",
				Source:   "dpd",
				Severity: model.ProblemSeverityError,
			},
			wantRequestURL: "https://example.invalid/dpd/bien%20compuesto",
			wantUserAgent:  "dlexa-test",
			wantAcceptLang: "es-ES,es;q=0.9,en;q=0.8",
		},
		{
			name: "non success transport handling becomes fetch problem",
			client: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(strings.NewReader("blocked")),
					Request:    req,
				}, nil
			}),
			wantProblem: &model.Problem{
				Code:     model.ProblemCodeDPDFetchFailed,
				Message:  "DPD request failed with status 403",
				Source:   "dpd",
				Severity: model.ProblemSeverityError,
			},
			wantRequestURL: "https://example.invalid/dpd/bien%20compuesto",
			wantUserAgent:  "dlexa-test",
			wantAcceptLang: "es-ES,es;q=0.9,en;q=0.8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewDPDFetcher("https://example.invalid/dpd", 2*time.Second, "dlexa-test")
			fetcher.Client = tt.client
			fetcher.now = func() time.Time { return fixedNow }

			document, err := fetcher.Fetch(context.Background(), request)
			if tt.wantProblem != nil {
				if err == nil {
					t.Fatal("Fetch() error = nil, want typed problem")
				}

				problem, ok := model.AsProblem(err)
				if !ok {
					t.Fatalf("Fetch() error = %T, want typed problem error", err)
				}

				if !reflect.DeepEqual(problem, *tt.wantProblem) {
					t.Fatalf("Fetch() problem = %#v, want %#v", problem, *tt.wantProblem)
				}

				if !reflect.DeepEqual(document, Document{}) {
					t.Fatalf("Fetch() document = %#v, want zero value on error", document)
				}
				return
			}

			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}

			if !reflect.DeepEqual(document, tt.wantDocument) {
				t.Fatalf("Fetch() document = %#v, want %#v", document, tt.wantDocument)
			}
		})
	}
}

func TestIsChallengeBodyEdgeCases(t *testing.T) {
	// Helpers
	padBytes := func(prefix []byte, total int) []byte {
		body := make([]byte, total)
		copy(body, prefix)
		for i := len(prefix); i < total; i++ {
			body[i] = 'x'
		}
		return body
	}

	challengeMarkers := []byte("Cloudflare challenge page detected")
	nonChallengeContent := []byte("normal html content without any markers")

	tests := []struct {
		name string
		body []byte
		want bool
	}{
		{
			name: "empty body",
			body: []byte{},
			want: false,
		},
		{
			name: "short body with challenge markers",
			body: []byte("<html>Cloudflare challenge</html>"),
			want: true,
		},
		{
			name: "short body without challenge markers",
			body: nonChallengeContent,
			want: false,
		},
		{
			name: "body exactly 1024 bytes with challenge markers at start",
			body: padBytes(challengeMarkers, 1024),
			want: true,
		},
		{
			name: "large body 500KB with challenge markers in first 1024 bytes",
			body: padBytes(challengeMarkers, 500*1024),
			want: true,
		},
		{
			name: "large body 500KB with challenge markers AFTER 1024 bytes",
			body: func() []byte {
				body := make([]byte, 500*1024)
				for i := 0; i < len(body); i++ {
					body[i] = 'x'
				}
				// Place markers well beyond 1024 bytes
				copy(body[2000:], challengeMarkers)
				return body
			}(),
			want: false, // After truncation fix, markers beyond 1024 are not seen
		},
		{
			name: "body with only cloudflare no challenge",
			body: []byte("Cloudflare is a CDN provider"),
			want: false,
		},
		{
			name: "body with only challenge no cloudflare",
			body: []byte("This is a challenge page"),
			want: false,
		},
		{
			name: "body with markers in mixed case",
			body: []byte("<html>CLOUDFLARE CHALLENGE</html>"),
			want: true,
		},
		{
			name: "non-ASCII UTF-8 body without markers",
			body: []byte("contenido en español con acentos: áéíóú ñ ¿¡ «»"),
			want: false,
		},
		{
			name: "non-ASCII UTF-8 body with markers among unicode",
			body: []byte("página de Cloudflare — verificación challenge activa «»"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isChallengeBody(tt.body)
			if got != tt.want {
				t.Fatalf("isChallengeBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}
