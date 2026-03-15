package doctor

import (
	"context"
	"time"
)

// NoopDoctor is a doctor.Service that always reports healthy.
type NoopDoctor struct{}

// NewNoopDoctor returns a NoopDoctor instance.
func NewNoopDoctor() *NoopDoctor {
	return &NoopDoctor{}
}

// Run returns a healthy report with a single bootstrap check.
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
