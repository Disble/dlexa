package render

import "fmt"

// StaticRegistry is a Registry backed by a fixed set of renderers.
type StaticRegistry struct {
	renderers map[string]Renderer
}

// NewRegistry creates a StaticRegistry from the given renderers.
func NewRegistry(renderers ...Renderer) *StaticRegistry {
	items := make(map[string]Renderer, len(renderers))
	for _, renderer := range renderers {
		items[renderer.Format()] = renderer
	}

	return &StaticRegistry{renderers: items}
}

// Renderer returns the renderer registered for the given format or an error if none exists.
func (r *StaticRegistry) Renderer(format string) (Renderer, error) {
	renderer, ok := r.renderers[format]
	if !ok {
		return nil, fmt.Errorf("renderer not registered for format %q", format)
	}

	return renderer, nil
}
