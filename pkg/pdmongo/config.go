package pdmongo

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/v2"
)

type Config struct {
	URI            string        `koanf:"uri"             mapstructure:"uri"             yaml:"uri"`
	Database       string        `koanf:"database"        mapstructure:"database"        yaml:"database"`
	ConnectTimeout time.Duration `koanf:"connect_timeout" mapstructure:"connect_timeout" yaml:"connect_timeout"`
	PingTimeout    time.Duration `koanf:"ping_timeout"    mapstructure:"ping_timeout"    yaml:"ping_timeout"`
}

func GetConfig(name string, k *koanf.Koanf) (*Config, error) {
	base := "mongo." + name

	var cfg Config
	if err := k.Unmarshal(base, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", base, err)
	}

	return &cfg, nil
}
