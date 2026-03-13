package normalize

import (
	"context"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type IdentityNormalizer struct{}

func NewIdentityNormalizer() *IdentityNormalizer {
	return &IdentityNormalizer{}
}

func (n *IdentityNormalizer) Normalize(ctx context.Context, descriptor model.SourceDescriptor, entries []model.Entry) ([]model.Entry, []model.Warning, error) {
	return n.NormalizeEntries(ctx, descriptor, entries)
}

func (n *IdentityNormalizer) NormalizeEntries(ctx context.Context, descriptor model.SourceDescriptor, entries []model.Entry) ([]model.Entry, []model.Warning, error) {
	_ = ctx
	normalized := make([]model.Entry, 0, len(entries))
	for _, entry := range entries {
		entry.Source = descriptor.Name
		if entry.Metadata == nil {
			entry.Metadata = map[string]string{}
		}
		entry.Metadata["normalized_by"] = "identity"
		normalized = append(normalized, entry)
	}

	warnings := []model.Warning{{
		Code:    "identity_normalizer",
		Message: "normalizer preserves parsed fields until canonical rules exist",
		Source:  descriptor.Name,
	}}

	return normalized, warnings, nil
}
