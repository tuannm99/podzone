package pdmongo

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	URI            string        `mapstructure:"uri"             yaml:"uri"`
	Database       string        `mapstructure:"database"        yaml:"database"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" yaml:"connect_timeout"`
	PingTimeout    time.Duration `mapstructure:"ping_timeout"    yaml:"ping_timeout"`
}

func GetConfigFromViper(name string, v *viper.Viper) (*Config, error) {
	base := "mongo." + name
	var cfg Config
	if sub := v.Sub(base); sub != nil {
		_ = sub.Unmarshal(&cfg)
	}
	return &cfg, nil
}
