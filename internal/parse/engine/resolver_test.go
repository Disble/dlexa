package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

func TestResolverRegistersAndResolvesArticleAndSearchParsers(t *testing.T) {
	resolver := NewResolver()
	article := articleParserFunc(func(ParseInput) (ArticleResult, []model.Warning, error) { return parse.Result{}, nil, nil })
	search := searchParserFunc(func(ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error) { return nil, nil, nil })

	resolver.RegisterArticle("dpd", article)
	resolver.RegisterSearch("search", search)

	gotArticle, err := resolver.ResolveArticle("dpd")
	if err != nil {
		t.Fatalf("ResolveArticle() error = %v", err)
	}
	if _, _, err := gotArticle.ParseArticle(ParseInput{}); err != nil {
		t.Fatalf("resolved article parser invocation error = %v", err)
	}

	gotSearch, err := resolver.ResolveSearch("search")
	if err != nil {
		t.Fatalf("ResolveSearch() error = %v", err)
	}
	if _, _, err := gotSearch.ParseSearch(ParseInput{}); err != nil {
		t.Fatalf("resolved search parser invocation error = %v", err)
	}
}

func TestResolverReturnsDeterministicErrorForMissingParser(t *testing.T) {
	resolver := NewResolver()

	_, articleErr := resolver.ResolveArticle("dpd")
	if articleErr == nil || !strings.Contains(articleErr.Error(), `article parser not registered for source "dpd"`) {
		t.Fatalf("ResolveArticle() error = %v, want deterministic missing article parser error", articleErr)
	}

	_, searchErr := resolver.ResolveSearch("search")
	if searchErr == nil || !strings.Contains(searchErr.Error(), `search parser not registered for source "search"`) {
		t.Fatalf("ResolveSearch() error = %v, want deterministic missing search parser error", searchErr)
	}
}

type articleParserFunc func(ParseInput) (ArticleResult, []model.Warning, error)

func (f articleParserFunc) ParseArticle(input ParseInput) (ArticleResult, []model.Warning, error) {
	return f(input)
}

type searchParserFunc func(ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error)

func (f searchParserFunc) ParseSearch(input ParseInput) ([]parse.ParsedSearchRecord, []model.Warning, error) {
	return f(input)
}

var _ = context.Background
