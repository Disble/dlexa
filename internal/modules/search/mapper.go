package search

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

func mapURLToCommand(query string, candidate model.SearchCandidate) (string, string, string) {
	if key := strings.TrimSpace(candidate.ArticleKey); key != "" && strings.TrimSpace(candidate.URL) == "" {
		return "dpd", key, fmt.Sprintf("dlexa dpd %s", key)
	}
	parsed, err := url.Parse(strings.TrimSpace(candidate.URL))
	if err != nil || parsed.Path == "" {
		return "unknown", strings.TrimSpace(candidate.ArticleKey), fmt.Sprintf("dlexa search %s", strings.TrimSpace(query))
	}
	cleanPath := strings.Trim(strings.TrimSpace(parsed.Path), "/")
	segments := strings.Split(cleanPath, "/")
	if len(segments) < 2 {
		return "unknown", path.Base(cleanPath), fmt.Sprintf("dlexa search %s", strings.TrimSpace(query))
	}
	moduleName := segments[0]
	slug := segments[len(segments)-1]
	switch moduleName {
	case "espanol-al-dia", "noticia", "duda-linguistica", "dpd":
		return moduleName, slug, fmt.Sprintf("dlexa %s %s", moduleName, slug)
	default:
		return "unknown", slug, fmt.Sprintf("dlexa search %s", strings.TrimSpace(query))
	}
}
