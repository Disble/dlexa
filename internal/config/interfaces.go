package config

import (
	"context"
	"time"
)

type DPDConfig struct {
	BaseURL   string
	Timeout   time.Duration
	UserAgent string
}

type RuntimeConfig struct {
	DefaultFormat  string
	DefaultSources []string
	CacheEnabled   bool
	DPD            DPDConfig
}

type Loader interface {
	Load(ctx context.Context) (RuntimeConfig, error)
}
