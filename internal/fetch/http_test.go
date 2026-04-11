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

	"github.com/Disble/dlexa/internal/model"
)

const (
	testBaseURL    = "https://example.invalid/dpd"
	testLookupURL  = "https://example.invalid/dpd/bien%20compuesto"
	testUserAgent  = "dlexa-test"
	testAcceptLang = "es-ES,es;q=0.9,en;q=0.8"
	testHTMLType   = "text/html; charset=utf-8"
)

func TestDPDFetcherClassifiesTransportOutcomesAndCapturesDocuments(t *testing.T) {
	fixedNow := time.Date(2026, time.March, 13, 16, 30, 0, 0, time.UTC)
	request := Request{Query: "  bien compuesto  ", Source: model.SourceDescriptor{Name: "dpd"}}

	tests := []struct {
		name         string
		client       Doer
		wantDocument Document
		wantProblem  *model.Problem
	}{
		{
			name: "successful html document capture",
			client: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if err := validateRequestHeaders(req); err != nil {
					return nil, err
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{testHTMLType},
					},
					Body:    io.NopCloser(strings.NewReader("<html>ok</html>")),
					Request: req,
				}, nil
			}),
			wantDocument: Document{
				URL:         testLookupURL,
				ContentType: testHTMLType,
				StatusCode:  http.StatusOK,
				Body:        []byte("<html>ok</html>"),
				RetrievedAt: fixedNow,
			},
		},
		{
			// When the DPD server redirects /dpd/solo → /dpd/tilde, the Go HTTP client
			// updates resp.Request.URL to the final (redirected) URL. Document.URL must
			// reflect the final URL (canonical page), while Document.RedirectedFrom
			// preserves the original lookup URL so the parser can derive the correct
			// query term ("bien compuesto", not whatever the redirect target is).
			name: "server redirect populates RedirectedFrom and URL reflects final target",
			client: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				redirectedURL, _ := req.URL.Parse("https://example.invalid/dpd/redirected-target")
				redirectedReq := req.Clone(req.Context())
				redirectedReq.URL = redirectedURL
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{testHTMLType},
					},
					Body:    io.NopCloser(strings.NewReader("<html>redirected content</html>")),
					Request: redirectedReq,
				}, nil
			}),
			wantDocument: Document{
				URL:            "https://example.invalid/dpd/redirected-target", // final URL after redirect
				RedirectedFrom: testLookupURL,                                   // original URL the user requested
				ContentType:    testHTMLType,
				StatusCode:     http.StatusOK,
				Body:           []byte("<html>redirected content</html>"),
				RetrievedAt:    fixedNow,
			},
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
		},
		{
			name:   "404 becomes not found problem",
			client: httpResponse(http.StatusNotFound, "missing"),
			wantProblem: &model.Problem{
				Code:     model.ProblemCodeDPDNotFound,
				Message:  "DPD entry not found for \"bien compuesto\"",
				Source:   "dpd",
				Severity: model.ProblemSeverityError,
			},
		},
		{
			name:   "challenge page becomes fetch problem",
			client: httpResponse(http.StatusForbidden, "<html><title>Just a moment...</title><div>Cloudflare challenge</div></html>"),
			wantProblem: &model.Problem{
				Code:     model.ProblemCodeDPDFetchFailed,
				Message:  "DPD request was challenged by upstream; browser-like profile still rejected",
				Source:   "dpd",
				Severity: model.ProblemSeverityError,
			},
		},
		{
			name:   "non success transport handling becomes fetch problem",
			client: httpResponse(http.StatusForbidden, "blocked"),
			wantProblem: &model.Problem{
				Code:     model.ProblemCodeDPDFetchFailed,
				Message:  "DPD request failed with status 403",
				Source:   "dpd",
				Severity: model.ProblemSeverityError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewDPDFetcher(testBaseURL, 2*time.Second, testUserAgent)
			fetcher.Client = tt.client
			fetcher.now = func() time.Time { return fixedNow }

			document, err := fetcher.Fetch(context.Background(), request)

			if tt.wantProblem != nil {
				assertFetchProblem(t, err, document, tt.wantProblem)
				return
			}

			assertFetchDocument(t, err, document, tt.wantDocument)
		})
	}
}

// validateRequestHeaders checks that the request carries the expected URL and headers.
// It is used inside round-trip closures that need to verify the outgoing request.
func validateRequestHeaders(req *http.Request) error {
	if got := req.URL.String(); got != testLookupURL {
		return errors.New("unexpected request URL: " + got)
	}

	if got := req.Header.Get("User-Agent"); got != testUserAgent {
		return errors.New("unexpected user agent: " + got)
	}

	if got := req.Header.Get("Accept-Language"); got != testAcceptLang {
		return errors.New("unexpected accept-language: " + got)
	}

	return nil
}

// assertFetchProblem is a test helper that verifies Fetch() returned the expected
// typed problem and a zero-value Document.
func assertFetchProblem(t *testing.T, err error, document Document, want *model.Problem) {
	t.Helper()

	if err == nil {
		t.Fatal("Fetch() error = nil, want typed problem")
	}

	problem, ok := model.AsProblem(err)
	if !ok {
		t.Fatalf("Fetch() error = %T, want typed problem error", err)
	}

	if !reflect.DeepEqual(problem, *want) {
		t.Fatalf("Fetch() problem = %#v, want %#v", problem, *want)
	}

	if !reflect.DeepEqual(document, Document{}) {
		t.Fatalf("Fetch() document = %#v, want zero value on error", document)
	}
}

// assertFetchDocument is a test helper that verifies Fetch() returned no error
// and a Document equal to want.
func assertFetchDocument(t *testing.T, err error, document Document, want Document) {
	t.Helper()

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if !reflect.DeepEqual(document, want) {
		t.Fatalf("Fetch() document = %#v, want %#v", document, want)
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

// httpResponse returns a Doer that always responds with the given status code and body.
func httpResponse(statusCode int, body string) roundTripFunc {
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}
