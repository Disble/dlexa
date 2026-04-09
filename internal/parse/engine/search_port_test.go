package engine

import (
	"context"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

func TestLegacySearchAdapterPreservesInputAndOutput(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "search", DisplayName: "RAE Search"}
	document := fetch.Document{URL: "https://example.invalid/search?q=solo", Body: []byte("body")}
	wantRecords := []parse.ParsedSearchRecord{{Title: "solo", URL: "https://www.rae.es/dpd/solo"}}
	wantWarnings := []model.Warning{{Code: "parse-warning", Source: descriptor.Name}}
	legacy := &legacySearchParserStub{records: wantRecords, warnings: wantWarnings}

	adapter := AdaptLegacySearchParser(legacy)
	records, warnings, err := adapter.ParseSearch(ParseInput{Ctx: context.Background(), Descriptor: descriptor, Document: document})
	if err != nil {
		t.Fatalf("ParseSearch() error = %v", err)
	}
	if !reflect.DeepEqual(records, wantRecords) {
		t.Fatalf("ParseSearch() records = %#v, want %#v", records, wantRecords)
	}
	if !reflect.DeepEqual(warnings, wantWarnings) {
		t.Fatalf("ParseSearch() warnings = %#v, want %#v", warnings, wantWarnings)
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

type legacySearchParserStub struct {
	ctx        context.Context
	descriptor model.SourceDescriptor
	document   fetch.Document
	records    []parse.ParsedSearchRecord
	warnings   []model.Warning
	err        error
}

func (s *legacySearchParserStub) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]parse.ParsedSearchRecord, []model.Warning, error) {
	s.ctx = ctx
	s.descriptor = descriptor
	s.document = document
	return s.records, s.warnings, s.err
}
