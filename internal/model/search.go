package model

import "time"

// SearchRequest holds the parameters for a DPD entry-search query.
type SearchRequest struct {
	Query   string
	Format  string
	NoCache bool
}

// SearchResult contains normalized DPD entry-search candidates and metadata.
type SearchResult struct {
	Request     SearchRequest
	Candidates  []SearchCandidate
	Warnings    []Warning
	Problems    []Problem
	CacheHit    bool
	GeneratedAt time.Time
}

// SearchCandidate is the format-neutral normalized contract for one DPD entry candidate.
type SearchCandidate struct {
	RawLabelHTML string `json:"raw_label_html"`
	DisplayText  string `json:"display_text"`
	ArticleKey   string `json:"article_key"`
}
