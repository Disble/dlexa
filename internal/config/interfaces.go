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

// SearchGovernanceConfig holds bounded upstream governance settings for search transports.
type SearchGovernanceConfig struct {
	CooldownBase      time.Duration
	CooldownMax       time.Duration
	RespectRetryAfter bool
}

// SearchConfig holds search-specific runtime configuration.
type SearchConfig struct {
	DefaultProviders []string
	MaxConcurrent    int
	Governance       SearchGovernanceConfig
}

// RuntimeConfig aggregates all runtime settings used by the application.
type RuntimeConfig struct {
	DefaultFormat        string
	DefaultLookupSources []string
	CacheEnabled         bool
	CacheTTL             time.Duration
	DPD                  DPDConfig
	Search               SearchConfig
}

// Loader loads and returns the runtime configuration.
type Loader interface {
	Load(ctx context.Context) (RuntimeConfig, error)
}
