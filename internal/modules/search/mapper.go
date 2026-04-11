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
		return moduleDPD, key, fmt.Sprintf("dlexa dpd %s", key)
	}
	parsed, err := url.Parse(strings.TrimSpace(candidate.URL))
	if err != nil || parsed.Path == "" {
		return moduleUnknown, strings.TrimSpace(candidate.ArticleKey), fmt.Sprintf(searchCommandFmt, strings.TrimSpace(query))
	}
	cleanPath := strings.Trim(strings.TrimSpace(parsed.Path), "/")
	segments := strings.Split(cleanPath, "/")
	if len(segments) < 2 {
		return moduleUnknown, path.Base(cleanPath), fmt.Sprintf(searchCommandFmt, strings.TrimSpace(query))
	}
	moduleName := segments[0]
	slug := segments[len(segments)-1]
	switch moduleName {
	case moduleEspanolAlDia, moduleNoticia, moduleDudaLinguistica, moduleDPD:
		if moduleName == moduleDPD {
			return moduleName, slug, fmt.Sprintf("dlexa dpd %s", slug)
		}
		return moduleName, slug, fmt.Sprintf("dlexa %s %s", moduleName, slug)
	default:
		return moduleUnknown, slug, fmt.Sprintf(searchCommandFmt, strings.TrimSpace(query))
	}
}
