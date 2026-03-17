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

func TestDPDSearchFetcherUsesKeysEndpointAndBrowserProfile(t *testing.T) {
	fixedNow := time.Date(2026, time.March, 17, 21, 5, 0, 0, time.UTC)
	request := Request{Query: " abu dhabi ", Source: model.SourceDescriptor{Name: "dpd"}}
	fetcher := NewDPDSearchFetcher("https://example.invalid/dpd", 2*time.Second, testUserAgent)
	fetcher.Client = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.URL.String(); got != "https://example.invalid/dpd/srv/keys?q=abu+dhabi" {
			return nil, errors.New("unexpected request URL: " + got)
		}
		if got := req.Header.Get("User-Agent"); got != testUserAgent {
			return nil, errors.New("unexpected user agent: " + got)
		}
		if got := req.Header.Get("Accept-Language"); got != testAcceptLang {
			return nil, errors.New("unexpected accept-language: " + got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Body:       io.NopCloser(strings.NewReader(`["<em>Abu Dhabi</em>|Abu Dabi"]`)),
			Request:    req,
		}, nil
	})
	fetcher.now = func() time.Time { return fixedNow }

	document, err := fetcher.Fetch(context.Background(), request)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	want := Document{
		URL:         "https://example.invalid/dpd/srv/keys?q=abu+dhabi",
		ContentType: "application/json; charset=utf-8",
		StatusCode:  http.StatusOK,
		Body:        []byte(`["<em>Abu Dhabi</em>|Abu Dabi"]`),
		RetrievedAt: fixedNow,
	}
	if !reflect.DeepEqual(document, want) {
		t.Fatalf("document = %#v, want %#v", document, want)
	}
}

func TestDPDSearchFetcherClassifiesChallengeAndTransportFailures(t *testing.T) {
	tests := []struct {
		name        string
		client      Doer
		wantProblem model.Problem
	}{
		{
			name: "challenge response",
			client: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusForbidden, Body: io.NopCloser(strings.NewReader("Cloudflare challenge")), Request: req}, nil
			}),
			wantProblem: model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "DPD entry search was challenged by upstream; browser-like profile still rejected", Source: "dpd", Severity: model.ProblemSeverityError},
		},
		{
			name: "transport failure",
			client: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return nil, errors.New("dial tcp refused")
			}),
			wantProblem: model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "fetch DPD entry search payload: dial tcp refused", Source: "dpd", Severity: model.ProblemSeverityError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewDPDSearchFetcher("https://example.invalid/dpd", time.Second, testUserAgent)
			fetcher.Client = tt.client

			document, err := fetcher.Fetch(context.Background(), Request{Query: "guion", Source: model.SourceDescriptor{Name: "dpd"}})
			assertFetchProblem(t, err, document, &tt.wantProblem)
		})
	}
}
