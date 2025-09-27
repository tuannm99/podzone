package pdlog

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func GetLogConfigFromViper(v *viper.Viper) (*Config, error) {
	var cfg Config
	cfg.Provider = "zap"
	cfg.Level = "info"
	cfg.Env = "dev"

	if sub := v.Sub("logger"); sub != nil {
		if err := sub.Unmarshal(&cfg); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func NewLogger(cfg *Config) (Logger, error) {
	switch cfg.Provider {
	case "slog":
		return NewSlogLogger(*cfg)
	case "zap", "":
		return NewZapLogger(*cfg)
	default:
		return nil, fmt.Errorf("unknown logger provider: %q", cfg.Provider)
	}
}

func Cleanup(logger Logger) {
	if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "sync /dev/stderr") {
		logger.Warn("logger sync failed", "error", err)
	}
}
