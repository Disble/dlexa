package config

import "context"

type StaticLoader struct {
	Config RuntimeConfig
}

func NewStaticLoader(cfg RuntimeConfig) *StaticLoader {
	return &StaticLoader{Config: cfg}
}

func (l *StaticLoader) Load(context.Context) (RuntimeConfig, error) {
	return l.Config, nil
}
