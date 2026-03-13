package render

import (
	"context"
	"encoding/json"

	"github.com/gentleman-programming/dlexa/internal/model"
)

type JSONRenderer struct{}

func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

func (r *JSONRenderer) Format() string {
	return "json"
}

func (r *JSONRenderer) Render(ctx context.Context, result model.LookupResult) ([]byte, error) {
	return r.RenderResult(ctx, result)
}

func (r *JSONRenderer) RenderResult(ctx context.Context, result model.LookupResult) ([]byte, error) {
	_ = ctx
	return json.MarshalIndent(result, "", "  ")
}
