// Package model defines the domain types shared across dlexa.
package model

import (
	"errors"
	"time"
)

// Problem codes, severity levels, and inline content kinds.
const (
	ProblemCodeSourceLookupFailed            = "source_lookup_failed"
	ProblemCodeSourcePipelineIncomplete      = "source_pipeline_incomplete"
	ProblemCodeDPDFetchFailed                = "dpd_fetch_failed"
	ProblemCodeDPDNotFound                   = "dpd_not_found"
	ProblemCodeDPDExtractFailed              = "dpd_extract_failed"
	ProblemCodeDPDTransformFailed            = "dpd_transform_failed"
	ProblemCodeDPDSearchFetchFailed          = "dpd_search_fetch_failed"
	ProblemCodeDPDSearchRateLimited          = "dpd_search_rate_limited"
	ProblemCodeDPDSearchParseFailed          = "dpd_search_parse_failed"
	ProblemCodeDPDSearchNormalizeFailed      = "dpd_search_normalize_failed"
	ProblemCodeSearchAllProvidersRateLimited = "search_all_providers_rate_limited"
	ProblemCodeArticleFetchFailed            = "article_fetch_failed"
	ProblemCodeArticleNotFound               = "article_not_found"
	ProblemCodeArticleExtractFailed          = "article_extract_failed"
	ProblemCodeArticleTransformFailed        = "article_transform_failed"

	// WarningCodeDPDRedirected is emitted when the DPD server transparently redirects
	// a lookup URL to a different entry (e.g. /dpd/solo → /dpd/tilde).
	// The redirect is an upstream decision; dlexa cannot override it, but informs
	// the user so the displayed content is not mistaken for the queried entry.
	WarningCodeDPDRedirected = "dpd_redirected"

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

	// DPD Signs - VALIDATED (real HTML evidence)
	InlineKindDigitalEdition       = "digital_edition"       // @ sign in <sup>@</sup>
	InlineKindConstructionMarker   = "construction_marker"   // + sign in <span class="nc">
	InlineKindBracketDefinition    = "bracket_definition"    // [...] in <dfn>
	InlineKindBracketPronunciation = "bracket_pronunciation" // [...] in <span class="nn">
	InlineKindBracketInterpolation = "bracket_interpolation" // [...] in <span class="yy">

	// WARNING: SPECULATIVE signs - no real HTML validation found in DPD articles.
	// Patterns are inferred from validated sign containers and MUST be revisited
	// when real examples are discovered.
	InlineKindAgrammatical = "agrammatical" // * sign (SPECULATIVE)
	InlineKindHypothetical = "hypothetical" // ‖ sign (SPECULATIVE)
	InlineKindPhoneme      = "phoneme"      // // sign (SPECULATIVE)

	// ARCHIVED SIGNS (not implemented):
	// < (etymology) and > (transformation) remain excluded because of HTML tag
	// collision risk. See testdata/dpd-signs-analysis/SIGN_ANALYSIS.md.
)

// LookupRequest holds the parameters for a dictionary lookup query.
type LookupRequest struct {
	Query   string
	Format  string
	Sources []string
	NoCache bool
}

// LookupResult contains the aggregated entries, structured misses, and metadata
// from a lookup.
type LookupResult struct {
	Request     LookupRequest
	Entries     []Entry
	Misses      []LookupMiss `json:"misses,omitempty"`
	Warnings    []Warning
	Problems    []Problem
	Sources     []SourceResult
	CacheHit    bool
	GeneratedAt time.Time
}

// LookupMissKind classifies a structured lookup miss outcome.
type LookupMissKind string

// LookupNextActionKind classifies an explicit next-step suggestion for a miss.
type LookupNextActionKind string

const (
	// LookupMissKindGenericNotFound marks a lookup miss with no native DPD suggestion.
	LookupMissKindGenericNotFound LookupMissKind = "generic_not_found"
	// LookupMissKindRelatedEntry marks a lookup miss with a native DPD related-entry suggestion.
	LookupMissKindRelatedEntry LookupMissKind = "related_entry"

	// LookupNextActionKindSearch marks an explicit recommendation to run the search command.
	LookupNextActionKindSearch LookupNextActionKind = "search"
)

// LookupMiss preserves format-neutral miss semantics across pipeline boundaries.
type LookupMiss struct {
	Kind       LookupMissKind    `json:"kind"`
	Query      string            `json:"query"`
	NoticeText string            `json:"notice_text,omitempty"`
	Suggestion *LookupSuggestion `json:"suggestion,omitempty"`
	NextAction *LookupNextAction `json:"next_action,omitempty"`
	Source     string            `json:"source,omitempty"`
}

// LookupSuggestion carries a native upstream suggestion for a near miss.
type LookupSuggestion struct {
	Kind         string `json:"kind"`
	DisplayText  string `json:"display_text"`
	EntryID      string `json:"entry_id,omitempty"`
	URL          string `json:"url,omitempty"`
	RawLabelHTML string `json:"raw_label_html,omitempty"`
}

// LookupNextAction carries an explicit user action recommendation for a miss.
type LookupNextAction struct {
	Kind    LookupNextActionKind `json:"kind"`
	Query   string               `json:"query"`
	Command string               `json:"command"`
}

// FallbackKind classifies the explicit fallback ladder kinds.
type FallbackKind string

const (
	// FallbackKindSyntax indicates the command shape was invalid.
	FallbackKindSyntax FallbackKind = "syntax"
	// FallbackKindNotFound indicates the command was valid but the content was not found.
	FallbackKindNotFound FallbackKind = "not_found"
	// FallbackKindRateLimited indicates the remote provider asked us to stop temporarily.
	FallbackKindRateLimited FallbackKind = "rate_limited"
	// FallbackKindUpstreamUnavailable indicates the remote provider is unavailable.
	FallbackKindUpstreamUnavailable FallbackKind = "upstream_unavailable"
	// FallbackKindParseFailure indicates upstream changed and parsing failed.
	FallbackKindParseFailure FallbackKind = "parse_failure"
)

// Envelope carries shared metadata for Markdown-wrapped module responses.
type Envelope struct {
	Module     string
	Title      string
	Source     string
	CacheState string
	Format     string
}

// HelpEnvelope carries agent-oriented Markdown help content.
type HelpEnvelope struct {
	Command     string
	Summary     string
	Syntax      string
	Examples    []string
	NextSteps   []string
	RecoveryTip string
}

// FallbackEnvelope carries structured fallback semantics independent from transport.
type FallbackEnvelope struct {
	Kind        FallbackKind `json:"kind"`
	Module      string       `json:"module,omitempty"`
	Title       string       `json:"title,omitempty"`
	Query       string       `json:"query,omitempty"`
	Message     string       `json:"message,omitempty"`
	Detail      string       `json:"detail,omitempty"`
	Syntax      string       `json:"syntax,omitempty"`
	Suggestion  string       `json:"suggestion,omitempty"`
	NextCommand string       `json:"next_command,omitempty"`
	Format      string       `json:"format,omitempty"`
}

// Entry represents a single dictionary entry with its content and metadata.
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

// Article is the structured representation of a dictionary article.
type Article struct {
	Dictionary   string
	Edition      string
	Lemma        string
	CanonicalURL string
	Sections     []Section
	Citation     Citation
}

// ArticleBlockKind identifies the type of content block within an article section.
type ArticleBlockKind string

// Supported article block kinds.
const (
	ArticleBlockKindParagraph ArticleBlockKind = "paragraph"
	ArticleBlockKindTable     ArticleBlockKind = "table"
)

// Section groups related blocks under a labeled heading within an article.
type Section struct {
	Label      string
	Title      string
	Blocks     []Block
	Paragraphs []Paragraph
	Children   []Section
}

// Block is a union container for a paragraph or table within a section.
type Block struct {
	Kind      ArticleBlockKind `json:"kind"`
	Paragraph *Paragraph       `json:"paragraph,omitempty"`
	Table     *Table           `json:"table,omitempty"`
}

// Table holds header and body rows for tabular content.
type Table struct {
	Headers []TableRow `json:"headers,omitempty"`
	Rows    []TableRow `json:"rows,omitempty"`
}

// TableRow is an ordered sequence of cells forming one row.
type TableRow struct {
	Cells []TableCell `json:"cells,omitempty"`
}

// TableCell represents a single cell with optional inline content and span attributes.
type TableCell struct {
	Text    string   `json:"text"`
	Inlines []Inline `json:"inlines,omitempty"`
	ColSpan int      `json:"colspan,omitempty"`
	RowSpan int      `json:"rowspan,omitempty"`
}

// Paragraph holds rendered markdown and its constituent inline elements.
type Paragraph struct {
	Markdown string
	Inlines  []Inline
}

// Inline is a span of styled or annotated text within a paragraph.
type Inline struct {
	Kind     string
	Variant  string
	Text     string
	Target   string
	Children []Inline
}

// Citation contains bibliographic reference data for an article.
type Citation struct {
	SourceLabel  string
	CanonicalURL string
	Edition      string
	ConsultedAt  string
	Text         string
}

// SourceDescriptor defines a dictionary source's identity and behavior.
type SourceDescriptor struct {
	Name        string
	DisplayName string
	Kind        string
	Priority    int
	Cacheable   bool
}

// SourceResult pairs a source descriptor with the entries or structured miss it
// produced.
type SourceResult struct {
	Source    SourceDescriptor
	Entries   []Entry
	Miss      *LookupMiss `json:"miss,omitempty"`
	Warnings  []Warning
	Problems  []Problem
	FetchedAt time.Time
}

// Warning represents a non-fatal issue encountered during lookup.
type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Source  string `json:"source,omitempty"`
}

// Problem describes a categorized error with source attribution and severity.
type Problem struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Source   string `json:"source,omitempty"`
	Severity string `json:"severity"`
}

// ProblemError wraps a Problem as a Go error for use in error chains.
type ProblemError struct {
	Problem Problem
	Err     error
}

// NewProblemError creates a ProblemError, defaulting severity to error if unset.
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

// AsProblem extracts a Problem from an error chain, if present.
func AsProblem(err error) (Problem, bool) {
	var problemErr *ProblemError
	if !errors.As(err, &problemErr) {
		return Problem{}, false
	}

	return problemErr.Problem, true
}
