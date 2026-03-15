// Package config defines configuration types and loaders for dlexa.
package config

import (
	"context"
	"time"
)

// DPDConfig holds connection settings for the RAE DPD source.
type DPDConfig struct {
	BaseURL   string
	Timeout   time.Duration
	UserAgent string
}

// RuntimeConfig aggregates all runtime settings used by the application.
type RuntimeConfig struct {
	DefaultFormat  string
	DefaultSources []string
	CacheEnabled   bool
	CacheTTL       time.Duration
	DPD            DPDConfig
}

// Loader loads and returns the runtime configuration.
type Loader interface {
	Load(ctx context.Context) (RuntimeConfig, error)
}
