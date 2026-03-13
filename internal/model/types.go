package model

import "time"

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
