package parse

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

var (
	reLiveSearchItem    = regexp.MustCompile(`(?is)<li>\s*<div class="row">.*?</li>`)
	reLiveSearchAnchor  = regexp.MustCompile(`(?is)<a[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	reLiveSearchSnippet = regexp.MustCompile(`(?is)<div class="pb-5">\s*<p>(.*?)</p>\s*</div>`)
	reLiveSearchTags    = regexp.MustCompile(`(?is)<[^>]+>`)
)

// LiveSearchParser extracts result cards from the general RAE search page.
type LiveSearchParser struct{}

// NewLiveSearchParser returns a ready-to-use live search parser.
func NewLiveSearchParser() *LiveSearchParser {
	return &LiveSearchParser{}
}

// Parse decodes the upstream search page into stable search records.
func (p *LiveSearchParser) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]ParsedSearchRecord, []model.Warning, error) {
	_ = ctx
	body := strings.TrimSpace(string(document.Body))
	if body == "" {
		return nil, nil, nil
	}

	items := reLiveSearchItem.FindAllString(body, -1)
	if len(items) == 0 {
		return nil, nil, nil
	}

	records := make([]ParsedSearchRecord, 0, len(items))
	for _, item := range items {
		record, ok := parseLiveSearchItem(document.URL, item)
		if !ok {
			continue
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, nil, model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchParseFailed, Message: fmt.Sprintf("extract live search candidates for %q", descriptor.Name), Source: descriptor.Name, Severity: model.ProblemSeverityError}, nil)
	}

	return records, nil, nil
}

func parseLiveSearchItem(baseURL, item string) (ParsedSearchRecord, bool) {
	anchor := reLiveSearchAnchor.FindStringSubmatch(item)
	if len(anchor) != 3 {
		return ParsedSearchRecord{}, false
	}
	snippet := reLiveSearchSnippet.FindStringSubmatch(item)
	resolvedURL, ok := resolveLiveSearchURL(baseURL, anchor[1])
	if !ok {
		return ParsedSearchRecord{}, false
	}
	record := ParsedSearchRecord{
		Title:   normalizeLiveSearchText(anchor[2]),
		Snippet: normalizeLiveSearchText(firstMatchGroup(snippet, 1)),
		URL:     resolvedURL,
	}
	if record.Title == "" || record.URL == "" {
		return ParsedSearchRecord{}, false
	}
	return record, true
}

func resolveLiveSearchURL(baseURL, href string) (string, bool) {
	href = strings.TrimSpace(href)
	if href == "" {
		return "", false
	}
	parsedHref, err := url.Parse(href)
	if err != nil {
		return "", false
	}
	if parsedHref.IsAbs() {
		return parsedHref.String(), true
	}
	parsedBase, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", false
	}
	return parsedBase.ResolveReference(parsedHref).String(), true
}

func normalizeLiveSearchText(raw string) string {
	text := html.UnescapeString(reLiveSearchTags.ReplaceAllString(strings.TrimSpace(raw), " "))
	return strings.Join(strings.Fields(text), " ")
}

func firstMatchGroup(matches []string, index int) string {
	if len(matches) <= index {
		return ""
	}
	return matches[index]
}
