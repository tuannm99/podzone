package pdhttp

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

type HttpConfig struct {
	Address        string   `mapstructure:"address"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

func NewHttpConfigFromKoanf(k *koanf.Koanf) (HttpConfig, error) {
	cfg := HttpConfig{Address: ":8000"}
	if k == nil {
		return cfg, fmt.Errorf("koanf is nil")
	}
	if err := k.Unmarshal("http", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal http config failed: %w", err)
	}
	if cfg.Address == "" {
		return cfg, fmt.Errorf("missing config: http.address")
	}
	return cfg, nil
}
