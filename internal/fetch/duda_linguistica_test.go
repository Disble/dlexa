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

func TestDudaLinguisticaFetcherBuildsArticleURLAndCapturesDocument(t *testing.T) {
	fixedNow := time.Date(2026, time.April, 9, 19, 30, 0, 0, time.UTC)
	fetcher := NewDudaLinguisticaFetcher("https://example.invalid/dpd", 2*time.Second, testUserAgent)
	fetcher.Client = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.URL.String(); got != "https://example.invalid/duda-linguistica/cuando-se-escriben-con-tilde-los-adverbios-en-mente" {
			return nil, errors.New("unexpected request URL: " + got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
			Body:       io.NopCloser(strings.NewReader("<html>ok</html>")),
			Request:    req,
		}, nil
	})
	fetcher.now = func() time.Time { return fixedNow }

	document, err := fetcher.Fetch(context.Background(), Request{Query: "cuando-se-escriben-con-tilde-los-adverbios-en-mente", Source: model.SourceDescriptor{Name: "duda-linguistica"}})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	want := Document{URL: "https://example.invalid/duda-linguistica/cuando-se-escriben-con-tilde-los-adverbios-en-mente", ContentType: "text/html; charset=utf-8", StatusCode: http.StatusOK, Body: []byte("<html>ok</html>"), RetrievedAt: fixedNow}
	if !reflect.DeepEqual(document, want) {
		t.Fatalf("document = %#v, want %#v", document, want)
	}
}

func TestDudaLinguisticaFetcherClassifiesNotFound(t *testing.T) {
	fetcher := NewDudaLinguisticaFetcher("https://example.invalid/dpd", time.Second, testUserAgent)
	fetcher.Client = httpResponse(http.StatusNotFound, "missing")

	document, err := fetcher.Fetch(context.Background(), Request{Query: "inexistente", Source: model.SourceDescriptor{Name: "duda-linguistica"}})
	assertFetchProblem(t, err, document, &model.Problem{Code: model.ProblemCodeArticleNotFound, Message: "article not found for \"inexistente\"", Source: "duda-linguistica", Severity: model.ProblemSeverityError})
}
