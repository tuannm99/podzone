package pdgraphql

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

type Config struct {
	Enabled    bool   `mapstructure:"enabled"`
	QueryPath  string `mapstructure:"query_path"`
	Playground struct {
		Enabled bool   `mapstructure:"enabled"`
		Path    string `mapstructure:"path"`
	} `mapstructure:"playground"`
}

func NewConfigFromKoanf(k *koanf.Koanf) (Config, error) {
	cfg := Config{
		Enabled:   false,
		QueryPath: "/query",
	}
	cfg.Playground.Enabled = true
	cfg.Playground.Path = "/"

	if k == nil {
		return cfg, fmt.Errorf("koanf is nil")
	}
	if err := k.Unmarshal("http.graphql", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal http.graphql failed: %w", err)
	}
	if cfg.Enabled && cfg.QueryPath == "" {
		return cfg, fmt.Errorf("missing config: http.graphql.query_path")
	}
	if cfg.Enabled && cfg.Playground.Enabled && cfg.Playground.Path == "" {
		return cfg, fmt.Errorf("missing config: http.graphql.playground.path")
	}
	return cfg, nil
}
