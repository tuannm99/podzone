// pkg/pdredis/config.go
package pdredis

import "github.com/spf13/viper"

type Config struct {
	URI      string `mapstructure:"uri"`
	Provider string `mapstructure:"provider"`
}

func GetConfigFromViper(name string, v *viper.Viper) (*Config, error) {
	base := "redis." + name
	var cfg Config
	if sub := v.Sub(base); sub != nil {
		_ = sub.Unmarshal(&cfg)
	}
	return &cfg, nil
}
