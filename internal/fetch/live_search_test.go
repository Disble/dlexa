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
	"github.com/Disble/dlexa/internal/testutil"
)

func TestLiveSearchFetcherUsesRAESearchPageRatherThanDPDKeys(t *testing.T) {
	fixedNow := time.Date(2026, time.April, 7, 23, 15, 0, 0, time.UTC)
	request := Request{Query: " " + testutil.LiveSearchQuery + " ", Source: model.SourceDescriptor{Name: "search"}}
	fetcher := NewLiveSearchFetcher(testBaseURL, 2*time.Second, testUserAgent)
	fetcher.Client = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotURL := req.URL.String()
		if gotURL != "https://example.invalid/search/node?keys=solo+o+solo" {
			return nil, errors.New("unexpected request URL: " + gotURL)
		}
		if strings.Contains(gotURL, "/srv/keys") {
			return nil, errors.New("live search must not call /srv/keys")
		}
		if got := req.Header.Get("User-Agent"); got != testUserAgent {
			return nil, errors.New("unexpected user agent: " + got)
		}
		if got := req.Header.Get("Accept"); !strings.Contains(got, "text/html") {
			return nil, errors.New("unexpected accept header: " + got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
			Body:       io.NopCloser(strings.NewReader("<html><body>ok</body></html>")),
			Request:    req,
		}, nil
	})
	fetcher.now = func() time.Time { return fixedNow }

	document, err := fetcher.Fetch(context.Background(), request)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	want := Document{
		URL:         "https://example.invalid/search/node?keys=solo+o+solo",
		ContentType: "text/html; charset=utf-8",
		StatusCode:  http.StatusOK,
		Body:        []byte("<html><body>ok</body></html>"),
		RetrievedAt: fixedNow,
	}
	if !reflect.DeepEqual(document, want) {
		t.Fatalf("document = %#v, want %#v", document, want)
	}
}

func TestLiveSearchFetcherClassifiesTransportFailures(t *testing.T) {
	fetcher := NewLiveSearchFetcher(testBaseURL, time.Second, testUserAgent)
	fetcher.Client = roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("dial tcp refused")
	})

	document, err := fetcher.Fetch(context.Background(), Request{Query: testutil.LiveSearchQuery, Source: model.SourceDescriptor{Name: "search"}})
	assertFetchProblem(t, err, document, &model.Problem{Code: model.ProblemCodeDPDSearchFetchFailed, Message: "fetch live search page: dial tcp refused", Source: "search", Severity: model.ProblemSeverityError})
}

func TestLiveSearchFetcherClassifiesRateLimitCooldownsExplicitly(t *testing.T) {
	now := time.Date(2026, time.April, 8, 19, 30, 0, 0, time.UTC)
	governed := NewGovernedDoer(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader("limited")), Request: req, Header: http.Header{"Retry-After": []string{"30"}}}, nil
	}), GovernanceConfig{CooldownBase: 5 * time.Second, CooldownMax: time.Minute, RespectRetryAfter: true})
	governed.now = func() time.Time { return now }
	fetcher := NewLiveSearchFetcher(testBaseURL, time.Second, testUserAgent)
	fetcher.Client = governed

	firstDocument, firstErr := fetcher.Fetch(context.Background(), Request{Query: testutil.LiveSearchQuery, Source: model.SourceDescriptor{Name: "search"}})
	assertFetchProblem(t, firstErr, firstDocument, &model.Problem{Code: model.ProblemCodeDPDSearchRateLimited, Message: "live search request was rate-limited by upstream (status 429)", Source: "search", Severity: model.ProblemSeverityError})

	now = now.Add(5 * time.Second)
	secondDocument, secondErr := fetcher.Fetch(context.Background(), Request{Query: testutil.LiveSearchQuery, Source: model.SourceDescriptor{Name: "search"}})
	assertFetchProblem(t, secondErr, secondDocument, &model.Problem{Code: model.ProblemCodeDPDSearchRateLimited, Message: "live search temporarily rate-limited: search transport cooling down for 25s after upstream rate limiting", Source: "search", Severity: model.ProblemSeverityError})
}
