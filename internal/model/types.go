package model

import (
	"errors"
	"time"
)

const (
	ProblemCodeSourceLookupFailed = "source_lookup_failed"
	ProblemCodeDPDFetchFailed     = "dpd_fetch_failed"
	ProblemCodeDPDNotFound        = "dpd_not_found"
	ProblemCodeDPDExtractFailed   = "dpd_extract_failed"
	ProblemCodeDPDTransformFailed = "dpd_transform_failed"

	ProblemSeverityError = "error"

	InlineKindText           = "text"
	InlineKindExample        = "example"
	InlineKindMention        = "mention"
	InlineKindLexicalHeading = "lexical_heading"
	InlineKindGloss          = "gloss"
	InlineKindCitationQuote  = "citation_quote"
	InlineKindBibliography   = "bibliography"
	InlineKindWorkTitle      = "work_title"
	InlineKindSmallCaps      = "small_caps"
	InlineKindEditorial      = "editorial_gloss"
	InlineKindExclusion      = "exclusion_marker"
	InlineKindScaffold       = "scaffold"
	InlineKindPattern        = "pattern"
	InlineKindCorrection     = "correction"
	InlineKindReference      = "reference"
	InlineKindEmphasis       = "emphasis"
)

type LookupRequest struct {
	Query   string
	Format  string
	Sources []string
	NoCache bool
}

type LookupResult struct {
	Request     LookupRequest
	Entries     []Entry
	Warnings    []Warning
	Problems    []Problem
	Sources     []SourceResult
	CacheHit    bool
	GeneratedAt time.Time
}

type Entry struct {
	ID       string
	Headword string
	Summary  string
	Content  string
	Source   string
	URL      string
	Metadata map[string]string
	Article  *Article
}

type Article struct {
	Dictionary   string
	Edition      string
	Lemma        string
	CanonicalURL string
	Sections     []Section
	Citation     Citation
}

type ArticleBlockKind string

const (
	ArticleBlockKindParagraph ArticleBlockKind = "paragraph"
	ArticleBlockKindTable     ArticleBlockKind = "table"
)

type Section struct {
	Label      string
	Title      string
	Blocks     []Block
	Paragraphs []Paragraph
	Children   []Section
}

type Block struct {
	Kind      ArticleBlockKind `json:"kind"`
	Paragraph *Paragraph       `json:"paragraph,omitempty"`
	Table     *Table           `json:"table,omitempty"`
}

type Table struct {
	Headers []TableRow `json:"headers,omitempty"`
	Rows    []TableRow `json:"rows,omitempty"`
}

type TableRow struct {
	Cells []TableCell `json:"cells,omitempty"`
}

type TableCell struct {
	Text string `json:"text"`
}

type Paragraph struct {
	Markdown string
	Inlines  []Inline
}

type Inline struct {
	Kind     string
	Variant  string
	Text     string
	Target   string
	Children []Inline
}

type Citation struct {
	SourceLabel  string
	CanonicalURL string
	Edition      string
	ConsultedAt  string
	Text         string
}

type SourceDescriptor struct {
	Name        string
	DisplayName string
	Kind        string
	Priority    int
	Cacheable   bool
}

type SourceResult struct {
	Source    SourceDescriptor
	Entries   []Entry
	Warnings  []Warning
	Problems  []Problem
	FetchedAt time.Time
}

type Warning struct {
	Code    string
	Message string
	Source  string
}

type Problem struct {
	Code     string
	Message  string
	Source   string
	Severity string
}

type ProblemError struct {
	Problem Problem
	Err     error
}

func NewProblemError(problem Problem, err error) *ProblemError {
	if problem.Severity == "" {
		problem.Severity = ProblemSeverityError
	}

	return &ProblemError{Problem: problem, Err: err}
}

func (e *ProblemError) Error() string {
	if e == nil {
		return ""
	}
	if e.Problem.Message != "" {
		return e.Problem.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Problem.Code
}

func (e *ProblemError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func AsProblem(err error) (Problem, bool) {
	var problemErr *ProblemError
	if !errors.As(err, &problemErr) {
		return Problem{}, false
	}

	return problemErr.Problem, true
}
