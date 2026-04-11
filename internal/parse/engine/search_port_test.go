package engine

import (
	"context"
	"reflect"
	"testing"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

const parseSearchErrFmt = "ParseSearch() error = %v"

func TestLegacySearchAdapterPreservesInputAndOutput(t *testing.T) {
	descriptor := model.SourceDescriptor{Name: "search", DisplayName: "RAE Search"}
	document := fetch.Document{URL: "https://example.invalid/search?q=solo", Body: []byte("body")}
	wantRecords := []parse.ParsedSearchRecord{{Title: "solo", URL: "https://www.rae.es/dpd/solo"}}
	wantWarnings := []model.Warning{{Code: "parse-warning", Source: descriptor.Name}}
	legacy := &legacySearchParserStub{records: wantRecords, warnings: wantWarnings}

	adapter := AdaptLegacySearchParser(legacy)
	records, warnings, err := adapter.ParseSearch(context.Background(), ParseInput{Descriptor: descriptor, Document: document})
	if err != nil {
		t.Fatalf(parseSearchErrFmt, err)
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
	if !legacy.calledWithContext {
		t.Fatal("legacy ctx = nil, want context passed through")
	}
}

func TestLiveSearchParserDelegatesToLegacyImplementation(t *testing.T) {
	parser := NewLiveSearchParser()
	descriptor := model.SourceDescriptor{Name: "search", DisplayName: "RAE Search"}
	document := fetch.Document{URL: "https://example.invalid/search?q=solo", Body: []byte(`<li><div class="row"><a href="/dpd/solo">solo</a><div class="pb-5"><p>Entrada DPD.</p></div></div></li>`)}

	records, warnings, err := parser.ParseSearch(context.Background(), ParseInput{Descriptor: descriptor, Document: document})
	if err != nil {
		t.Fatalf(parseSearchErrFmt, err)
	}
	if len(warnings) != 0 {
		t.Fatalf("ParseSearch() warnings = %#v, want none", warnings)
	}
	if len(records) != 1 || records[0].Title != "solo" || records[0].URL != "https://example.invalid/dpd/solo" {
		t.Fatalf("ParseSearch() records = %#v, want delegated live-search parsing result", records)
	}
}

func TestDPDSearchParserDelegatesToLegacyImplementation(t *testing.T) {
	parser := NewDPDSearchParser()
	descriptor := model.SourceDescriptor{Name: "dpd", DisplayName: "DPD"}
	document := fetch.Document{Body: []byte(`["solo|solo"]`)}

	records, warnings, err := parser.ParseSearch(context.Background(), ParseInput{Descriptor: descriptor, Document: document})
	if err != nil {
		t.Fatalf(parseSearchErrFmt, err)
	}
	if len(warnings) != 0 {
		t.Fatalf("ParseSearch() warnings = %#v, want none", warnings)
	}
	if len(records) != 1 || records[0].RawLabelHTML != "solo" || records[0].ArticleKey != "solo" {
		t.Fatalf("ParseSearch() records = %#v, want delegated DPD-search parsing result", records)
	}
}

type legacySearchParserStub struct {
	calledWithContext bool
	descriptor        model.SourceDescriptor
	document          fetch.Document
	records           []parse.ParsedSearchRecord
	warnings          []model.Warning
	err               error
}

func (s *legacySearchParserStub) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]parse.ParsedSearchRecord, []model.Warning, error) {
	s.calledWithContext = ctx != nil
	s.descriptor = descriptor
	s.document = document
	return s.records, s.warnings, s.err
}
