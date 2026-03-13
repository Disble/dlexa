package source

import (
	"fmt"
	"sort"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type StaticRegistry struct {
	sources []Source
}

func NewStaticRegistry(sources ...Source) *StaticRegistry {
	ordered := append([]Source(nil), sources...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Descriptor().Priority < ordered[j].Descriptor().Priority
	})

	return &StaticRegistry{sources: ordered}
}

func (r *StaticRegistry) SourcesFor(request model.LookupRequest) ([]Source, error) {
	if len(request.Sources) == 0 {
		return r.sources, nil
	}

	allowed := make(map[string]struct{}, len(request.Sources))
	for _, name := range request.Sources {
		allowed[name] = struct{}{}
	}

	selected := make([]Source, 0, len(request.Sources))
	for _, item := range r.sources {
		if _, ok := allowed[item.Descriptor().Name]; ok {
			selected = append(selected, item)
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no sources matched request: %v", request.Sources)
	}

	return selected, nil
}
