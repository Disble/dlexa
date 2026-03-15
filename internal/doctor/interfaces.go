// Package doctor provides health-check diagnostics for dlexa sources.
package doctor

import (
	"context"
	"time"
)

// Check represents the result of a single diagnostic check.
type Check struct {
	Name   string
	Status string
	Detail string
}

// Report aggregates all diagnostic checks into a single health report.
type Report struct {
	Healthy     bool
	Checks      []Check
	GeneratedAt time.Time
}

// Service runs health-check diagnostics and returns a report.
type Service interface {
	Run(ctx context.Context) (Report, error)
}
