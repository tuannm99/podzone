package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/tuannm99/podzone/pkg/pdauthn"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type RPCConfig struct {
	GRPCHost string `mapstructure:"grpc_host"`
	GRPCPort string `mapstructure:"grpc_port"`
}

type ServerConfig struct {
	Authn          pdauthn.Config
	AppRedirectURL string
	Auth           RPCConfig `mapstructure:"auth"`
}

func NewServerConfig(k *koanf.Koanf) ServerConfig {
	cfg := ServerConfig{
		Authn: pdauthn.Config{
			JWTSecret: toolkit.GetEnv("JWT_SECRET", ""),
			JWTKey:    toolkit.GetEnv("JWT_KEY", ""),
		},
		AppRedirectURL: toolkit.GetEnv("APP_REDIRECT_URL", ""),
	}
	if k != nil {
		if cfg.Authn.JWTSecret == "" {
			cfg.Authn.JWTSecret = k.String("iam.authn.jwt_secret")
		}
		if cfg.Authn.JWTSecret == "" {
			cfg.Authn.JWTSecret = k.String("auth.jwt_secret")
		}
		if cfg.Authn.JWTKey == "" {
			cfg.Authn.JWTKey = k.String("iam.authn.jwt_key")
		}
		if cfg.Authn.JWTKey == "" {
			cfg.Authn.JWTKey = k.String("auth.jwt_key")
		}
		if cfg.AppRedirectURL == "" {
			cfg.AppRedirectURL = k.String("iam.app_redirect_url")
		}
		if cfg.AppRedirectURL == "" {
			cfg.AppRedirectURL = k.String("auth.app_redirect_url")
		}
		if cfg.Auth.GRPCHost == "" {
			cfg.Auth.GRPCHost = k.String("iam.auth.grpc_host")
		}
		if cfg.Auth.GRPCPort == "" {
			cfg.Auth.GRPCPort = k.String("iam.auth.grpc_port")
		}
	}
	if cfg.Auth.GRPCHost == "" {
		cfg.Auth.GRPCHost = toolkit.GetEnv("AUTH_GRPC_HOST", "localhost")
	}
	if cfg.Auth.GRPCPort == "" {
		cfg.Auth.GRPCPort = toolkit.GetEnv("AUTH_GRPC_PORT", "50052")
	}
	return cfg
}
