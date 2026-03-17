package render

import "fmt"

// StaticSearchRegistry is a SearchRendererResolver backed by a fixed set of search renderers.
type StaticSearchRegistry struct {
	renderers map[string]SearchRenderer
}

// NewSearchRegistry creates a StaticSearchRegistry from the given renderers.
func NewSearchRegistry(renderers ...SearchRenderer) *StaticSearchRegistry {
	items := make(map[string]SearchRenderer, len(renderers))
	for _, renderer := range renderers {
		items[renderer.Format()] = renderer
	}
	return &StaticSearchRegistry{renderers: items}
}

// Renderer returns the search renderer registered for the given format.
func (r *StaticSearchRegistry) Renderer(format string) (SearchRenderer, error) {
	renderer, ok := r.renderers[format]
	if !ok {
		return nil, fmt.Errorf("search renderer not registered for format %q", format)
	}
	return renderer, nil
}
