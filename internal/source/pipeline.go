package source

import (
	"context"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/normalize"
	"github.com/Disble/dlexa/internal/parse"
)

// PipelineSource implements Source by chaining a fetcher, parser, and normalizer.
type PipelineSource struct {
	descriptor model.SourceDescriptor
	fetcher    fetch.Fetcher
	parser     parse.Parser
	normalizer normalize.Normalizer
}

// NewPipelineSource creates a PipelineSource with the given adapters.
func NewPipelineSource(
	descriptor model.SourceDescriptor,
	fetcher fetch.Fetcher,
	parser parse.Parser,
	normalizer normalize.Normalizer,
) *PipelineSource {
	return &PipelineSource{
		descriptor: descriptor,
		fetcher:    fetcher,
		parser:     parser,
		normalizer: normalizer,
	}
}

// Descriptor returns the source metadata for this pipeline.
func (s *PipelineSource) Descriptor() model.SourceDescriptor {
	return s.descriptor
}

// Lookup fetches, parses, and normalizes a dictionary entry for the given request.
func (s *PipelineSource) Lookup(ctx context.Context, request model.LookupRequest) (model.SourceResult, error) {
	result := model.SourceResult{Source: s.descriptor}

	if s.fetcher == nil || s.parser == nil || s.normalizer == nil {
		result.Problems = append(result.Problems, model.Problem{
			Code:     "source_pipeline_incomplete",
			Message:  "source pipeline is missing one or more adapters",
			Source:   s.descriptor.Name,
			Severity: "error",
		})
		return result, nil
	}

	document, err := s.fetcher.Fetch(ctx, fetch.Request{Query: request.Query, Source: s.descriptor})
	if err != nil {
		return result, err
	}

	parsedResult, parseWarnings, err := s.parser.Parse(ctx, s.descriptor, document)
	if err != nil {
		return result, err
	}

	normalizedResult, err := s.normalizer.Normalize(ctx, s.descriptor, parsedResult)
	if err != nil {
		return result, err
	}

	result.Entries = normalizedResult.Entries
	result.Miss = normalizedResult.Miss
	result.Warnings = append(result.Warnings, parseWarnings...)
	result.Warnings = append(result.Warnings, normalizedResult.Warnings...)
	result.FetchedAt = document.RetrievedAt

	return result, nil
}
