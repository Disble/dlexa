// Package search adapts semantic search behavior to the shared module contract.
package search

import (
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

func curateCandidates(query string, candidates []model.SearchCandidate) []model.SearchCandidate {
	curated := make([]model.SearchCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if shouldDropCandidate(candidate) {
			continue
		}
		curated = append(curated, enrichCandidate(query, candidate))
	}
	return curated
}

func shouldDropCandidate(candidate model.SearchCandidate) bool {
	url := strings.TrimSpace(candidate.URL)
	if strings.Contains(url, "/institucion/") {
		return true
	}
	if strings.Contains(url, "/noticia/") {
		return !isRescuedNoticia(candidate)
	}
	return false
}

func enrichCandidate(query string, candidate model.SearchCandidate) model.SearchCandidate {
	title := firstNonEmpty(candidate.Title, candidate.DisplayText, candidate.ArticleKey)
	candidate.Title = title
	if strings.TrimSpace(candidate.Snippet) == "" {
		candidate.Snippet = strings.TrimSpace(candidate.DisplayText)
	}
	if strings.TrimSpace(candidate.Classification) == "" {
		candidate.Classification = classifyCandidate(candidate)
	}
	moduleName, id, nextCommand := mapURLToCommand(query, candidate)
	candidate.Module = moduleName
	candidate.ID = id
	candidate.NextCommand = nextCommand
	candidate.Deferred = moduleName != "dpd" && moduleName != "unknown"
	if strings.TrimSpace(candidate.SourceHint) == "" {
		candidate.SourceHint = firstNonEmpty(candidate.SourceHint, "RAE")
	}
	return candidate
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func classifyCandidate(candidate model.SearchCandidate) string {
	url := strings.TrimSpace(candidate.URL)
	if strings.Contains(url, "/noticia/") && isRescuedNoticia(candidate) {
		return "faq"
	}
	if strings.Contains(url, "/espanol-al-dia/") {
		return "linguistic-article"
	}
	if strings.Contains(url, "/dpd/") {
		return "dpd-entry"
	}
	if candidate.ArticleKey != "" {
		return "dpd-entry"
	}
	return "unknown"
}

func isRescuedNoticia(candidate model.SearchCandidate) bool {
	title := strings.ToLower(strings.TrimSpace(candidate.Title))
	return strings.HasPrefix(title, strings.ToLower("Preguntas frecuentes:")) || strings.Contains(title, "tilde") || strings.Contains(title, "normativa")
}
