package doctor

import (
	"context"
	"time"
)

type NoopDoctor struct{}

func NewNoopDoctor() *NoopDoctor {
	return &NoopDoctor{}
}

func (d *NoopDoctor) Run(context.Context) (Report, error) {
	return Report{
		Healthy: true,
		Checks: []Check{{
			Name:   "bootstrap",
			Status: "ok",
			Detail: "doctor wiring is ready; concrete checks can be added per platform",
		}},
		GeneratedAt: time.Now().UTC(),
	}, nil
}
