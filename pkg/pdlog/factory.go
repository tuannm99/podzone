package pdlog

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/v2"
)

func GetLogConfig(k *koanf.Koanf) (*Config, error) {
	cfg := Config{
		Provider: "zap",
		Level:    "info",
		Env:      "dev",
	}

	// Merge overrides from config
	if k != nil && k.Exists("logger") {
		if err := k.Unmarshal("logger", &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal logger config: %w", err)
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
