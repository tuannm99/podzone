package config

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

type Config struct {
	Auth RPCConfig `mapstructure:"auth"`
	IAM  RPCConfig `mapstructure:"iam"`
}

type RPCConfig struct {
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
	if err := k.Unmarshal("partner", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal partner config failed: %w", err)
	}
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = k.String("partner.auth.jwt_secret")
	}
	if cfg.Auth.JWTKey == "" {
		cfg.Auth.JWTKey = k.String("partner.auth.jwt_key")
	}
	if cfg.Auth.GRPCHost == "" {
		cfg.Auth.GRPCHost = k.String("partner.auth.grpc_host")
	}
	if cfg.Auth.GRPCPort == "" {
		cfg.Auth.GRPCPort = k.String("partner.auth.grpc_port")
	}
	if cfg.IAM.JWTSecret == "" {
		cfg.IAM.JWTSecret = k.String("partner.iam.jwt_secret")
	}
	if cfg.IAM.JWTKey == "" {
		cfg.IAM.JWTKey = k.String("partner.iam.jwt_key")
	}
	if cfg.IAM.GRPCHost == "" {
		cfg.IAM.GRPCHost = k.String("partner.iam.grpc_host")
	}
	if cfg.IAM.GRPCPort == "" {
		cfg.IAM.GRPCPort = k.String("partner.iam.grpc_port")
	}
	if cfg.Auth.JWTSecret == "" {
		return cfg, fmt.Errorf("missing config: partner.auth.jwt_secret")
	}
	if cfg.Auth.GRPCHost == "" {
		cfg.Auth.GRPCHost = "localhost"
	}
	if cfg.Auth.GRPCPort == "" {
		cfg.Auth.GRPCPort = "50051"
	}
	if cfg.IAM.GRPCHost == "" {
		cfg.IAM.GRPCHost = cfg.Auth.GRPCHost
	}
	if cfg.IAM.GRPCPort == "" {
		cfg.IAM.GRPCPort = "50053"
	}
	return cfg, nil
}
