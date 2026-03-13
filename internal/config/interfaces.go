package config

import "context"

type RuntimeConfig struct {
	DefaultFormat  string
	DefaultSources []string
	CacheEnabled   bool
}

type Loader interface {
	Load(ctx context.Context) (RuntimeConfig, error)
}
