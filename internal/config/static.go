package config

import (
	"context"
	"time"
)

// DefaultDPDBaseURL is the base URL for the RAE DPD dictionary.
const (
	DefaultDPDBaseURL   = "https://www.rae.es/dpd"
	DefaultDPDUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	DefaultDPDTimeout   = 10 * time.Second
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
		DefaultFormat:  "markdown",
		DefaultSources: []string{"dpd"},
		CacheEnabled:   true,
		CacheTTL:       24 * time.Hour,
		DPD: DPDConfig{
			BaseURL:   DefaultDPDBaseURL,
			Timeout:   DefaultDPDTimeout,
			UserAgent: DefaultDPDUserAgent,
		},
	}
}

// Load returns the embedded configuration unchanged.
func (l *StaticLoader) Load(context.Context) (RuntimeConfig, error) {
	return l.Config, nil
}
