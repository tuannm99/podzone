package pdelasticsearch

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/v2"
)

// Config holds Elasticsearch connection settings
type Config struct {
	Addresses      []string      `koanf:"addresses"       yaml:"addresses"       mapstructure:"addresses"`
	Username       string        `koanf:"username"        yaml:"username"        mapstructure:"username"`
	Password       string        `koanf:"password"        yaml:"password"        mapstructure:"password"`
	ConnectTimeout time.Duration `koanf:"connect_timeout" yaml:"connect_timeout" mapstructure:"connect_timeout"`
	PingTimeout    time.Duration `koanf:"ping_timeout"    yaml:"ping_timeout"    mapstructure:"ping_timeout"`
}

func GetConfig(name string, k *koanf.Koanf) (*Config, error) {
	base := "elasticsearch." + name

	var cfg Config
	if err := k.Unmarshal(base, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", base, err)
	}

	// Defaults
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}
	if cfg.PingTimeout == 0 {
		cfg.PingTimeout = 3 * time.Second
	}

	return &cfg, nil
}
