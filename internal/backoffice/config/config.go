package config

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

type Config struct {
	Auth AuthConfig `mapstructure:"auth"`
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
	JWTKey    string `mapstructure:"jwt_key"`
	GRPCHost  string `mapstructure:"grpc_host"`
	GRPCPort  string `mapstructure:"grpc_port"`
}

func NewConfigFromKoanf(k *koanf.Koanf) (Config, error) {
	cfg := Config{}
	if k == nil {
		return cfg, fmt.Errorf("koanf is nil")
	}
	if err := k.Unmarshal("backoffice", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal backoffice config failed: %w", err)
	}
	if cfg.Auth.JWTSecret == "" {
		return cfg, fmt.Errorf("missing config: backoffice.auth.jwt_secret")
	}
	if cfg.Auth.GRPCHost == "" {
		cfg.Auth.GRPCHost = "localhost"
	}
	if cfg.Auth.GRPCPort == "" {
		cfg.Auth.GRPCPort = "50051"
	}
	return cfg, nil
}
