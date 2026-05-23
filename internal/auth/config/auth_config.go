package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type RPCConfig struct {
	GRPCHost string `mapstructure:"grpc_host"`
	GRPCPort string `mapstructure:"grpc_port"`
}

type AuthConfig struct {
	JWTSecret      string
	JWTKey         string
	AppRedirectURL string
	IAM            RPCConfig `mapstructure:"iam"`
}

func NewAuthConfig(k *koanf.Koanf) AuthConfig {
	cfg := AuthConfig{
		JWTSecret:      toolkit.GetEnv("JWT_SECRET", ""),
		JWTKey:         toolkit.GetEnv("JWT_KEY", ""),
		AppRedirectURL: toolkit.GetEnv("APP_REDIRECT_URL", ""),
	}
	if k != nil {
		if cfg.JWTSecret == "" {
			cfg.JWTSecret = k.String("auth.jwt_secret")
		}
		if cfg.JWTKey == "" {
			cfg.JWTKey = k.String("auth.jwt_key")
		}
		if cfg.AppRedirectURL == "" {
			cfg.AppRedirectURL = k.String("auth.app_redirect_url")
		}
		if cfg.IAM.GRPCHost == "" {
			cfg.IAM.GRPCHost = k.String("auth.iam.grpc_host")
		}
		if cfg.IAM.GRPCPort == "" {
			cfg.IAM.GRPCPort = k.String("auth.iam.grpc_port")
		}
	}
	if cfg.IAM.GRPCHost == "" {
		cfg.IAM.GRPCHost = toolkit.GetEnv("IAM_GRPC_HOST", "localhost")
	}
	if cfg.IAM.GRPCPort == "" {
		cfg.IAM.GRPCPort = toolkit.GetEnv("IAM_GRPC_PORT", "50053")
	}
	return cfg
}
