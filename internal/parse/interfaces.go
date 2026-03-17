// Package parse provides HTML and markdown document parsers for dictionary sources.
package parse

import (
	"context"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

// Parser converts a fetched document into structured parsed articles.
type Parser interface {
	Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (Result, []model.Warning, error)
}

// Result holds the articles extracted from a parsed document.
type Result struct {
	Articles []ParsedArticle
	Miss     *ParsedLookupMiss
}

// ParsedLookupMissKind classifies a structured DPD lookup miss.
type ParsedLookupMissKind string

const (
	// ParsedLookupMissKindGenericNotFound marks a miss page with no native related-entry suggestion.
	ParsedLookupMissKindGenericNotFound ParsedLookupMissKind = "generic_not_found"
	// ParsedLookupMissKindRelatedEntry marks a miss page that includes a native related-entry suggestion.
	ParsedLookupMissKindRelatedEntry ParsedLookupMissKind = "related_entry"
)

// ParsedLookupMiss holds raw miss facts extracted from the DPD response.
type ParsedLookupMiss struct {
	Kind         ParsedLookupMissKind
	Query        string
	NoticeText   string
	RelatedEntry *ParsedRelatedEntry
}

// ParsedRelatedEntry holds the upstream related-entry suggestion extracted from a DPD miss page.
type ParsedRelatedEntry struct {
	RawLabelHTML string
	DisplayText  string
	EntryID      string
	Href         string
}

// ParsedBlockKind discriminates block content types within a section.
type ParsedBlockKind string

// Supported ParsedBlockKind values.
const (
	ParsedBlockKindParagraph ParsedBlockKind = "paragraph"
	ParsedBlockKindTable     ParsedBlockKind = "table"
)

// ParsedArticle represents a single dictionary article with its metadata and content sections.
type ParsedArticle struct {
	Dictionary   string
	Edition      string
	EntryID      string
	Lemma        string
	CanonicalURL string
	Sections     []ParsedSection
	Citation     ParsedCitation
}

// ParsedSection is a numbered or labelled subdivision of an article.
type ParsedSection struct {
	Label      string
	Title      string
	Blocks     []ParsedBlock
	Paragraphs []ParsedParagraph
	Children   []ParsedSection
}

// ParsedBlock is a content block that is either a paragraph or a table.
type ParsedBlock struct {
	Kind      ParsedBlockKind
	Paragraph *ParsedParagraph
	Table     *ParsedTable
}

// ParsedTable holds the header and body rows of an HTML table.
type ParsedTable struct {
	Headers []ParsedTableRow
	Rows    []ParsedTableRow
}

// ParsedTableRow is a single row of cells within a parsed table.
type ParsedTableRow struct {
	Cells []ParsedTableCell
}

// ParsedTableCell is a single cell with its rendered HTML, inline tree, and optional span attributes.
type ParsedTableCell struct {
	HTML    string
	Inlines []model.Inline
	ColSpan int
	RowSpan int
}

// ParsedParagraph holds the normalised HTML and inline semantic tree for a paragraph.
type ParsedParagraph struct {
	HTML    string
	Inlines []model.Inline
}

// ParsedCitation carries the citation block in both raw HTML and plain-text forms.
type ParsedCitation struct {
	HTML string
	Text string
}
