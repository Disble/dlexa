package normalize

import (
	"context"
	"fmt"
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
		if title == "" || url == "" {
			continue
		}
		candidates = append(candidates, model.SearchCandidate{
			Title:       title,
			DisplayText: title,
			Snippet:     strings.TrimSpace(record.Snippet),
			SourceHint:  "RAE",
			URL:         url,
		})
	}

	if len(records) > 0 && len(candidates) == 0 {
		return nil, nil, model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchNormalizeFailed, Message: fmt.Sprintf("normalize live search candidates for %q", descriptor.Name), Source: descriptor.Name, Severity: model.ProblemSeverityError}, nil)
	}

	return candidates, nil, nil
}
