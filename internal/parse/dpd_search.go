package parse

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

// ParsedSearchRecord is the structural parse contract for one upstream keys item.
type ParsedSearchRecord struct {
	RawLabelHTML string
	ArticleKey   string
	Title        string
	Snippet      string
	URL          string
}

// DPDSearchParser decodes the remote /dpd/srv/keys JSON payload and splits first-pipe records.
type DPDSearchParser struct{}

// NewDPDSearchParser returns a ready-to-use DPD search parser.
func NewDPDSearchParser() *DPDSearchParser {
	return &DPDSearchParser{}
}

// Parse decodes the upstream JSON array and extracts valid display|article_key records.
func (p *DPDSearchParser) Parse(ctx context.Context, descriptor model.SourceDescriptor, document fetch.Document) ([]ParsedSearchRecord, []model.Warning, error) {
	_ = ctx
	var rawItems []string
	if err := json.Unmarshal(document.Body, &rawItems); err != nil {
		return nil, nil, model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchParseFailed, Message: fmt.Sprintf("decode DPD entry search payload: %v", err), Source: descriptor.Name, Severity: model.ProblemSeverityError}, err)
	}

	records := make([]ParsedSearchRecord, 0, len(rawItems))
	for _, item := range rawItems {
		rawLabel, articleKey, ok := splitSearchRecord(item)
		if !ok {
			continue
		}
		records = append(records, ParsedSearchRecord{RawLabelHTML: rawLabel, ArticleKey: articleKey})
	}

	if len(rawItems) > 0 && len(records) == 0 {
		return nil, nil, model.NewProblemError(model.Problem{Code: model.ProblemCodeDPDSearchParseFailed, Message: "DPD entry search payload contained no usable display|article_key records", Source: descriptor.Name, Severity: model.ProblemSeverityError}, nil)
	}

	return records, nil, nil
}

func splitSearchRecord(item string) (string, string, bool) {
	parts := strings.SplitN(item, "|", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	rawLabel := strings.TrimSpace(parts[0])
	articleKey := strings.TrimSpace(parts[1])
	if rawLabel == "" || articleKey == "" {
		return "", "", false
	}
	return rawLabel, articleKey, true
}
