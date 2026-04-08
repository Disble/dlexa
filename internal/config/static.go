package config

import (
	"context"
	"time"
)

// DefaultDPDBaseURL is the base URL for the RAE DPD dictionary.
const (
	DefaultDPDBaseURL          = "https://www.rae.es/dpd"
	DefaultDPDUserAgent        = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	DefaultDPDTimeout          = 10 * time.Second
	DefaultSearchMaxConcurrent = 2
	DefaultSearchCooldownBase  = 5 * time.Second
	DefaultSearchCooldownMax   = 1 * time.Minute
)

// StaticLoader returns a fixed RuntimeConfig without external I/O.
type StaticLoader struct {
	Config RuntimeConfig
}

// NewStaticLoader creates a StaticLoader from the given configuration.
func NewStaticLoader(cfg RuntimeConfig) *StaticLoader {
	return &StaticLoader{Config: cfg}
}

// DefaultRuntimeConfig returns production-ready defaults for all settings.
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		DefaultFormat:        "markdown",
		DefaultLookupSources: []string{"dpd"},
		CacheEnabled:         true,
		CacheTTL:             24 * time.Hour,
		DPD: DPDConfig{
			BaseURL:   DefaultDPDBaseURL,
			Timeout:   DefaultDPDTimeout,
			UserAgent: DefaultDPDUserAgent,
		},
		Search: SearchConfig{
			DefaultProviders: []string{"search", "dpd"},
			MaxConcurrent:    DefaultSearchMaxConcurrent,
			Governance: SearchGovernanceConfig{
				CooldownBase:      DefaultSearchCooldownBase,
				CooldownMax:       DefaultSearchCooldownMax,
				RespectRetryAfter: true,
			},
		},
	}
}

// Load returns the embedded configuration unchanged.
func (l *StaticLoader) Load(context.Context) (RuntimeConfig, error) {
	return l.Config, nil
}
