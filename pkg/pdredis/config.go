package pdredis

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

type Config struct {
	URI string `koanf:"uri" mapstructure:"uri"`
}

func GetConfigFromKoanf(name string, k *koanf.Koanf) (*Config, error) {
	base := "redis." + name

	var cfg Config
	if err := k.Unmarshal(base, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", base, err)
	}
	return &cfg, nil
}
