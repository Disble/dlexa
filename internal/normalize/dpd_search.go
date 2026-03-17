package normalize

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

var (
	reSearchTags      = regexp.MustCompile(`(?is)<[^>]+>`)
	reSearchVariant   = regexp.MustCompile(`(?is)<span class="vers">(.*?)</span>`)
	reSearchSup       = regexp.MustCompile(`(?is)<sup[^>]*>(.*?)</sup>`)
	reSearchExclusion = regexp.MustCompile(`⊗([\pL\pN])`)
)

// DPDSearchNormalizer converts parsed search records into format-neutral search candidates.
type DPDSearchNormalizer struct{}

// NewDPDSearchNormalizer returns a ready-to-use DPD search normalizer.
func NewDPDSearchNormalizer() *DPDSearchNormalizer {
	return &DPDSearchNormalizer{}
}

// Normalize converts parsed search records into normalized search candidates.
func (n *DPDSearchNormalizer) Normalize(ctx context.Context, descriptor model.SourceDescriptor, records []parse.ParsedSearchRecord) ([]model.SearchCandidate, []model.Warning, error) {
	_ = ctx
	candidates := make([]model.SearchCandidate, 0, len(records))
	for _, record := range records {
		display := flattenSearchLabel(record.RawLabelHTML)
		if display == "" {
			continue
		}
		candidates = append(candidates, model.SearchCandidate{RawLabelHTML: record.RawLabelHTML, DisplayText: display, ArticleKey: record.ArticleKey})
	}

	if len(records) > 0 && len(candidates) == 0 {
		return nil, nil, model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchNormalizeFailed, Message: fmt.Sprintf("normalize DPD entry search candidates for %q", descriptor.Name), Source: descriptor.Name, Severity: model.ProblemSeverityError}, nil)
	}

	return candidates, nil, nil
}

func flattenSearchLabel(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	variant := reSearchVariant.MatchString(raw)
	raw = reSearchVariant.ReplaceAllString(raw, "$1")
	raw = reSearchSup.ReplaceAllString(raw, "$1")
	text := html.UnescapeString(reSearchTags.ReplaceAllString(raw, ""))
	text = strings.Join(strings.Fields(text), " ")
	text = reSearchExclusion.ReplaceAllString(text, "⊗ $1")
	text = strings.TrimSpace(text)
	if variant && !strings.HasPrefix(text, "var. ") {
		text = "var. " + text
	}
	return text
}
