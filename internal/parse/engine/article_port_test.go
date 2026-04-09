// Package engine tests the parser-engine foundation contracts.
package engine

import (
	"context"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

func TestLegacyArticleAdapterPreservesInputAndOutput(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "dpd", DisplayName: "DPD"}
	document := fetch.Document{URL: "https://example.invalid/dpd/bien", Body: []byte("body")}
	wantResult := parse.Result{Articles: []parse.ParsedArticle{{Lemma: "bien"}}}
	wantWarnings := []model.Warning{{Code: "parse-warning", Source: descriptor.Name}}
	legacy := &legacyArticleParserStub{result: wantResult, warnings: wantWarnings}

	adapter := AdaptLegacyArticleParser(legacy)
	result, warnings, err := adapter.ParseArticle(ParseInput{Ctx: context.Background(), Descriptor: descriptor, Document: document})
	if err != nil {
		t.Fatalf("ParseArticle() error = %v", err)
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Fatalf("ParseArticle() result = %#v, want %#v", result, wantResult)
	}
	if !reflect.DeepEqual(warnings, wantWarnings) {
		t.Fatalf("ParseArticle() warnings = %#v, want %#v", warnings, wantWarnings)
	}
	if !reflect.DeepEqual(legacy.descriptor, descriptor) {
		t.Fatalf("legacy descriptor = %#v, want %#v", legacy.descriptor, descriptor)
	}
	if !reflect.DeepEqual(legacy.document, document) {
		t.Fatalf("legacy document = %#v, want %#v", legacy.document, document)
	}
	if legacy.ctx == nil {
		t.Fatal("legacy ctx = nil, want context passed through")
	}
}

type legacyArticleParserStub struct {
	ctx        context.Context
	descriptor model.SourceDescriptor
	document   fetch.Document
	result     parse.Result
	warnings   []model.Warning
	err        error
}

func (s *legacyArticleParserStub) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (parse.Result, []model.Warning, error) {
	s.ctx = ctx
	s.descriptor = descriptor
	s.document = document
	return s.result, s.warnings, s.err
}
