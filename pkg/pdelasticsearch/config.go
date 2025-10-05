package pdelasticsearch

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds Elasticsearch connection settings
type Config struct {
	Addresses      []string      `mapstructure:"addresses"       yaml:"addresses"`
	Username       string        `mapstructure:"username"        yaml:"username"`
	Password       string        `mapstructure:"password"        yaml:"password"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" yaml:"connect_timeout"`
	PingTimeout    time.Duration `mapstructure:"ping_timeout"    yaml:"ping_timeout"`
}

func GetConfigFromViper(name string, v *viper.Viper) (*Config, error) {
	base := "elasticsearch." + name
	var cfg Config
	if sub := v.Sub(base); sub != nil {
		_ = sub.Unmarshal(&cfg)
	}
	if len(cfg.Addresses) == 0 {
		cfg.Addresses = []string{"http://localhost:9200"}
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}
	if cfg.PingTimeout == 0 {
		cfg.PingTimeout = 3 * time.Second
	}
	return &cfg, nil
}

