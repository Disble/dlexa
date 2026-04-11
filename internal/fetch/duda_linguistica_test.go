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

const dudaLinguisticaPath = "duda-linguistica"

func TestDudaLinguisticaFetcherBuildsArticleURLAndCapturesDocument(t *testing.T) {
	fixedNow := time.Date(2026, time.April, 9, 19, 30, 0, 0, time.UTC)
	fetcher := NewDudaLinguisticaFetcher(testBaseURL, 2*time.Second, testUserAgent)
	fetcher.Client = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.URL.String(); got != "https://example.invalid/"+dudaLinguisticaPath+"/cuando-se-escriben-con-tilde-los-adverbios-en-mente" {
			return nil, errors.New("unexpected request URL: " + got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{testHTMLType}},
			Body:       io.NopCloser(strings.NewReader("<html>ok</html>")),
			Request:    req,
		}, nil
	})
	fetcher.now = func() time.Time { return fixedNow }

	document, err := fetcher.Fetch(context.Background(), Request{Query: "cuando-se-escriben-con-tilde-los-adverbios-en-mente", Source: model.SourceDescriptor{Name: dudaLinguisticaPath}})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	want := Document{URL: "https://example.invalid/" + dudaLinguisticaPath + "/cuando-se-escriben-con-tilde-los-adverbios-en-mente", ContentType: testHTMLType, StatusCode: http.StatusOK, Body: []byte("<html>ok</html>"), RetrievedAt: fixedNow}
	if !reflect.DeepEqual(document, want) {
		t.Fatalf("document = %#v, want %#v", document, want)
	}
}

func TestDudaLinguisticaFetcherClassifiesNotFound(t *testing.T) {
	fetcher := NewDudaLinguisticaFetcher(testBaseURL, time.Second, testUserAgent)
	fetcher.Client = httpResponse(http.StatusNotFound, "missing")

	document, err := fetcher.Fetch(context.Background(), Request{Query: "inexistente", Source: model.SourceDescriptor{Name: dudaLinguisticaPath}})
	assertFetchProblem(t, err, document, &model.Problem{Code: model.ProblemCodeArticleNotFound, Message: "article not found for \"inexistente\"", Source: dudaLinguisticaPath, Severity: model.ProblemSeverityError})
}
