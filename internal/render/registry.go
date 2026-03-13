package render

import "fmt"

type StaticRegistry struct {
	renderers map[string]Renderer
}

func NewRegistry(renderers ...Renderer) *StaticRegistry {
	items := make(map[string]Renderer, len(renderers))
	for _, renderer := range renderers {
		items[renderer.Format()] = renderer
	}

	return &StaticRegistry{renderers: items}
}

func (r *StaticRegistry) Renderer(format string) (Renderer, error) {
	renderer, ok := r.renderers[format]
	if !ok {
		return nil, fmt.Errorf("renderer not registered for format %q", format)
	}

	return renderer, nil
}
