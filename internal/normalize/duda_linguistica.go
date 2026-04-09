package normalize

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/parse"
)

// DudaLinguisticaNormalizer converts parsed duda-linguistica articles into lookup entries.
type DudaLinguisticaNormalizer struct{}

// NewDudaLinguisticaNormalizer returns a new normalizer for duda-linguistica articles.
func NewDudaLinguisticaNormalizer() *DudaLinguisticaNormalizer {
	return &DudaLinguisticaNormalizer{}
}

// Normalize converts parsed articles into canonical lookup entries.
func (n *DudaLinguisticaNormalizer) Normalize(ctx context.Context, descriptor model.SourceDescriptor, result parse.Result) (Result, error) {
	_ = ctx
	entries := make([]model.Entry, 0, len(result.Articles))
	for _, article := range result.Articles {
		if len(article.Sections) == 0 {
			return Result{}, model.NewProblemError(model.Problem{
				Code:     model.ProblemCodeArticleTransformFailed,
				Message:  fmt.Sprintf("duda-linguistica article has no sections to normalize for %q", descriptor.Name),
				Source:   descriptor.Name,
				Severity: model.ProblemSeverityError,
			}, nil)
		}

		normalizedArticle := model.Article{
			Dictionary:   article.Dictionary,
			Edition:      article.Edition,
			Lemma:        article.Lemma,
			CanonicalURL: article.CanonicalURL,
			Sections:     normalizeSections(article.Sections),
			Citation:     normalizeCitation(article),
		}

		entryID := strings.TrimSpace(article.EntryID)
		if entryID == "" {
			entryID = slug(article.Lemma)
		}
		entries = append(entries, model.Entry{
			ID:       fmt.Sprintf("%s:%s#%s", descriptor.Name, slug(article.Lemma), slug(entryID)),
			Headword: article.Lemma,
			Summary:  article.Dictionary,
			Content:  markdownBody(normalizedArticle),
			Source:   descriptor.Name,
			URL:      article.CanonicalURL,
			Article:  &normalizedArticle,
			Metadata: map[string]string{"normalized_by": descriptor.Name, "entry_id": entryID},
		})
	}
	return Result{Entries: entries}, nil
}
