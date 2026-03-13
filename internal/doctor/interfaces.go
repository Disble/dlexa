package doctor

import (
	"context"
	"time"
)

type Check struct {
	Name   string
	Status string
	Detail string
}

type Report struct {
	Healthy     bool
	Checks      []Check
	GeneratedAt time.Time
}

type Service interface {
	Run(ctx context.Context) (Report, error)
}
