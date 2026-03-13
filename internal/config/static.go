package config

import (
	"context"
	"time"
)

const (
	DefaultDPDBaseURL   = "https://www.rae.es/dpd"
	DefaultDPDUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	DefaultDPDTimeout   = 10 * time.Second
)

type StaticLoader struct {
	Config RuntimeConfig
}

func NewStaticLoader(cfg RuntimeConfig) *StaticLoader {
	return &StaticLoader{Config: cfg}
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		DefaultFormat:  "markdown",
		DefaultSources: []string{"dpd"},
		CacheEnabled:   true,
		DPD: DPDConfig{
			BaseURL:   DefaultDPDBaseURL,
			Timeout:   DefaultDPDTimeout,
			UserAgent: DefaultDPDUserAgent,
		},
	}
}

func (l *StaticLoader) Load(context.Context) (RuntimeConfig, error) {
	return l.Config, nil
}
