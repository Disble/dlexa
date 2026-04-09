package engine

import (
	"fmt"
	"strings"
)

// Resolver stores parser registrations per family and source.
type Resolver struct {
	articles map[string]ArticleParser
	search   map[string]SearchParser
}

// NewResolver creates an empty parser resolver.
func NewResolver() *Resolver {
	return &Resolver{
		articles: make(map[string]ArticleParser),
		search:   make(map[string]SearchParser),
	}
}

// RegisterArticle registers an article parser for a source.
func (r *Resolver) RegisterArticle(source string, parser ArticleParser) {
	if r == nil {
		return
	}
	r.articles[strings.TrimSpace(source)] = parser
}

// RegisterSearch registers a search parser for a source.
func (r *Resolver) RegisterSearch(source string, parser SearchParser) {
	if r == nil {
		return
	}
	r.search[strings.TrimSpace(source)] = parser
}

// ResolveArticle resolves an article parser by source.
func (r *Resolver) ResolveArticle(source string) (ArticleParser, error) {
	if r == nil {
		return nil, fmt.Errorf("article parser not registered for source %q", strings.TrimSpace(source))
	}
	parser, ok := r.articles[strings.TrimSpace(source)]
	if !ok || parser == nil {
		return nil, fmt.Errorf("article parser not registered for source %q", strings.TrimSpace(source))
	}
	return parser, nil
}

// ResolveSearch resolves a search parser by source.
func (r *Resolver) ResolveSearch(source string) (SearchParser, error) {
	if r == nil {
		return nil, fmt.Errorf("search parser not registered for source %q", strings.TrimSpace(source))
	}
	parser, ok := r.search[strings.TrimSpace(source)]
	if !ok || parser == nil {
		return nil, fmt.Errorf("search parser not registered for source %q", strings.TrimSpace(source))
	}
	return parser, nil
}
