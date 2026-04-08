// Package search adapts semantic search behavior to the shared module contract.
package search

import (
	"net/url"
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/Disble/dlexa/internal/model"
)

func curateCandidates(query string, candidates []model.SearchCandidate) []model.SearchCandidate {
	curated := make([]rankedCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if shouldDropCandidate(candidate) {
			continue
		}
		enriched := enrichCandidate(query, candidate)
		curated = append(curated, rankedCandidate{candidate: enriched, score: candidateRankScore(query, enriched)})
	}
	if len(curated) == 0 {
		return nil
	}
	sort.SliceStable(curated, func(i, j int) bool {
		if curated[i].score != curated[j].score {
			return curated[i].score > curated[j].score
		}
		return curated[i].candidate.Title < curated[j].candidate.Title
	})
	return deduplicateCandidates(curated)
}

type rankedCandidate struct {
	candidate model.SearchCandidate
	score     int
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

func candidateRankScore(query string, candidate model.SearchCandidate) int {
	score := classificationRank(candidate.Classification) * 100
	if strings.TrimSpace(candidate.ArticleKey) != "" {
		score += 20
	}
	score += queryAffinityScore(query, candidate)
	if strings.TrimSpace(candidate.Snippet) != "" {
		score += 10
	}
	if strings.TrimSpace(candidate.URL) != "" {
		score++
	}
	return score
}

func queryAffinityScore(query string, candidate model.SearchCandidate) int {
	normalizedQuery := normalizeSearchText(query)
	if normalizedQuery == "" {
		return 0
	}

	canonicalLabel := firstNonEmpty(candidate.Title, candidate.DisplayText, candidate.ArticleKey, candidate.ID)
	normalizedLabel := normalizeSearchText(canonicalLabel)
	if normalizedLabel == "" {
		return 0
	}

	if normalizedLabel == normalizedQuery {
		return 140
	}

	isShortQuery := len(normalizedQuery) <= 3
	if isShortQuery {
		return 0
	}

	queryTokens := tokenizeNormalized(query)
	labelTokens := tokenizeNormalized(canonicalLabel)
	labelSet := make(map[string]struct{}, len(labelTokens))
	for _, token := range labelTokens {
		labelSet[token] = struct{}{}
	}

	matched := 0
	for _, token := range queryTokens {
		if _, ok := labelSet[token]; ok {
			matched++
		}
	}
	if matched > 0 {
		return matched * 15
	}
	if strings.Contains(normalizedLabel, normalizedQuery) {
		return 60
	}

	return 0
}

func tokenizeNormalized(s string) []string {
	normalized := normalizeSearchText(s)
	if normalized == "" {
		return nil
	}

	return strings.FieldsFunc(normalized, func(r rune) bool {
		return unicode.IsSpace(r) || r == '-'
	})
}

func classificationRank(classification string) int {
	switch strings.TrimSpace(classification) {
	case "dpd-entry":
		return 4
	case "faq":
		return 3
	case "linguistic-article":
		return 2
	default:
		return 1
	}
}

func deduplicateCandidates(ranked []rankedCandidate) []model.SearchCandidate {
	seen := make(map[string]struct{}, len(ranked))
	curated := make([]model.SearchCandidate, 0, len(ranked))
	for _, item := range ranked {
		key := candidateDedupKey(item.candidate)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		curated = append(curated, item.candidate)
	}
	return curated
}

func candidateDedupKey(candidate model.SearchCandidate) string {
	if key := canonicalDPDTarget(candidate); key != "" {
		return "dpd:" + key
	}
	for _, value := range []string{candidate.NextCommand, candidate.URL, candidate.ArticleKey, candidate.Title} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return "unknown"
}

func canonicalDPDTarget(candidate model.SearchCandidate) string {
	if key := strings.TrimSpace(candidate.ArticleKey); key != "" {
		return normalizeDPDTarget(key)
	}
	parsed, err := url.Parse(strings.TrimSpace(candidate.URL))
	if err != nil {
		return ""
	}
	cleanPath := strings.Trim(strings.TrimSpace(parsed.Path), "/")
	if !strings.HasPrefix(cleanPath, "dpd/") {
		return ""
	}
	return normalizeDPDTarget(path.Base(cleanPath))
}

func normalizeDPDTarget(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "_", " ")
	return normalizeSearchText(value)
}

func normalizeSearchText(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			builder.WriteRune(stripAccent(r))
		case unicode.IsSpace(r), r == '-', r == '_', r == '/':
			builder.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func stripAccent(r rune) rune {
	switch r {
	case 'á', 'à', 'ä', 'â':
		return 'a'
	case 'é', 'è', 'ë', 'ê':
		return 'e'
	case 'í', 'ì', 'ï', 'î':
		return 'i'
	case 'ó', 'ò', 'ö', 'ô':
		return 'o'
	case 'ú', 'ù', 'ü', 'û':
		return 'u'
	default:
		return r
	}
}
