package normalize

import (
	"context"
	"strings"
	"testing"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

func TestDPDSearchNormalizerBuildsFormatNeutralCandidates(t *testing.T) {
	normalizer := NewDPDSearchNormalizer()
	candidates, warnings, err := normalizer.Normalize(context.Background(), model.SourceDescriptor{Name: "dpd"}, []parse.ParsedSearchRecord{{RawLabelHTML: `<em>Abu Dhabi</em>`, ArticleKey: "Abu Dabi"}, {RawLabelHTML: `<span class="bolaspa">⊗</span>solo <span class="nc">(o</span> solamente<span class="nc">)</span> hacer que <span class="nc">+ infinitivo</span>`, ArticleKey: "hacer"}, {RawLabelHTML: `<span class="vers">guion<sup>2</sup></span>`, ArticleKey: "guion"}})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if len(candidates) != 3 {
		t.Fatalf("candidates = %#v", candidates)
	}
	if got := candidates[0]; got.RawLabelHTML != `<em>Abu Dhabi</em>` || got.DisplayText != "Abu Dhabi" || got.ArticleKey != "Abu Dabi" {
		t.Fatalf("first candidate = %#v", got)
	}
	if got := candidates[1].DisplayText; got != "⊗ solo (o solamente) hacer que + infinitivo" {
		t.Fatalf("rejected display = %q", got)
	}
	if got := candidates[2].DisplayText; !strings.Contains(got, "var.") || !strings.Contains(got, "guion2") {
		t.Fatalf("variant display = %q, want visible variant + superscript cue", got)
	}
	for _, candidate := range candidates {
		if strings.Contains(candidate.DisplayText, "->") {
			t.Fatalf("display_text leaked renderer token: %#v", candidate)
		}
	}
}
