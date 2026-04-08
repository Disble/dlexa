package normalize

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// LiveSearchNormalizer converts parsed live-search records into curated candidates.
type LiveSearchNormalizer struct{}

// NewLiveSearchNormalizer returns a ready-to-use live search normalizer.
func NewLiveSearchNormalizer() *LiveSearchNormalizer {
	return &LiveSearchNormalizer{}
}

// Normalize converts parsed live search records into normalized search candidates.
func (n *LiveSearchNormalizer) Normalize(ctx context.Context, descriptor model.SourceDescriptor, records []parse.ParsedSearchRecord) ([]model.SearchCandidate, []model.Warning, error) {
	_ = ctx
	candidates := make([]model.SearchCandidate, 0, len(records))
	for _, record := range records {
		title := strings.TrimSpace(record.Title)
		url := strings.TrimSpace(record.URL)
		if title == "" || url == "" || !matchesSearchSurface(descriptor.Name, url) {
			continue
		}
		candidates = append(candidates, model.SearchCandidate{
			Title:       title,
			DisplayText: title,
			Snippet:     strings.TrimSpace(record.Snippet),
			SourceHint:  sourceHintForDescriptor(descriptor.Name),
			URL:         url,
		})
	}

	if len(records) > 0 && len(candidates) == 0 {
		return nil, nil, model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchNormalizeFailed, Message: fmt.Sprintf("normalize live search candidates for %q", descriptor.Name), Source: descriptor.Name, Severity: model.ProblemSeverityError}, nil)
	}

	return candidates, nil, nil
}

func matchesSearchSurface(providerName, rawURL string) bool {
	providerName = strings.TrimSpace(providerName)
	if providerName == "" || providerName == "search" {
		return true
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	return strings.Contains(strings.Trim(parsed.Path, "/"), providerName)
}

func sourceHintForDescriptor(providerName string) string {
	if strings.TrimSpace(providerName) == "search" {
		return "RAE"
	}
	return strings.TrimSpace(providerName)
}
