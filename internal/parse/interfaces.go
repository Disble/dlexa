package parse

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/fetch"
	"github.com/gentleman-programming/dlexa/internal/model"
)

type Parser interface {
	Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) (Result, []model.Warning, error)
}

type Result struct {
	Articles []ParsedArticle
}

type ParsedBlockKind string

const (
	ParsedBlockKindParagraph ParsedBlockKind = "paragraph"
	ParsedBlockKindTable     ParsedBlockKind = "table"
)

type ParsedArticle struct {
	Dictionary   string
	Edition      string
	EntryID      string
	Lemma        string
	CanonicalURL string
	Sections     []ParsedSection
	Citation     ParsedCitation
}

type ParsedSection struct {
	Label      string
	Title      string
	Blocks     []ParsedBlock
	Paragraphs []ParsedParagraph
	Children   []ParsedSection
}

type ParsedBlock struct {
	Kind      ParsedBlockKind
	Paragraph *ParsedParagraph
	Table     *ParsedTable
}

type ParsedTable struct {
	Headers []ParsedTableRow
	Rows    []ParsedTableRow
}

type ParsedTableRow struct {
	Cells []ParsedTableCell
}

type ParsedTableCell struct {
	HTML    string
	Inlines []model.Inline
	ColSpan int
	RowSpan int
}

type ParsedParagraph struct {
	HTML    string
	Inlines []model.Inline
}

type ParsedCitation struct {
	HTML string
	Text string
}
