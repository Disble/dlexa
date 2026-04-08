package model

import "time"

// SearchOutcome classifies the final semantic-search outcome after curation.
type SearchOutcome string

const (
	// SearchOutcomeResults indicates the search completed with visible curated candidates.
	SearchOutcomeResults SearchOutcome = "results"
	// SearchOutcomeNoResults indicates the search succeeded but no curated candidates remained.
	SearchOutcomeNoResults SearchOutcome = "no_results"
)

// SearchRequest holds the parameters for a DPD entry-search query.
type SearchRequest struct {
	Query   string
	Format  string
	Sources []string
	NoCache bool
}

// SearchResult contains normalized DPD entry-search candidates and metadata.
type SearchResult struct {
	Request     SearchRequest
	Candidates  []SearchCandidate
	Outcome     SearchOutcome
	Warnings    []Warning
	Problems    []Problem
	CacheHit    bool
	GeneratedAt time.Time
}

// SearchCandidate is the format-neutral normalized contract for one DPD entry candidate.
type SearchCandidate struct {
	RawLabelHTML   string `json:"raw_label_html"`
	DisplayText    string `json:"display_text"`
	ArticleKey     string `json:"article_key"`
	Title          string `json:"title,omitempty"`
	Snippet        string `json:"snippet,omitempty"`
	Module         string `json:"module,omitempty"`
	ID             string `json:"id,omitempty"`
	NextCommand    string `json:"next_command,omitempty"`
	Deferred       bool   `json:"deferred"`
	Classification string `json:"classification,omitempty"`
	SourceHint     string `json:"source_hint,omitempty"`
	URL            string `json:"url,omitempty"`
}
