package config

import (
	"strings"

	"github.com/knadh/koanf/v2"

	"github.com/tuannm99/podzone/pkg/pdauthn"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type IAMConfig struct {
	GRPCHost string `koanf:"grpc_host"`
	GRPCPort string `koanf:"grpc_port"`
}

type AuthConfig struct {
	Authn                  pdauthn.Config
	IAM                    IAMConfig
	ServiceToken           string
	BackofficeURL          string
	BackofficeServiceToken string
}

func NewAuthConfig(k *koanf.Koanf) AuthConfig {
	cfg := AuthConfig{
		Authn: pdauthn.Config{
			JWTSecret: toolkit.GetEnv("JWT_SECRET", ""),
			JWTKey:    toolkit.GetEnv("JWT_KEY", ""),
		},
		IAM: IAMConfig{
			GRPCHost: toolkit.GetEnv("IAM_GRPC_HOST", "iam-service"),
			GRPCPort: toolkit.GetEnv("IAM_GRPC_PORT", "50053"),
		},
		ServiceToken:           toolkit.GetEnv("ONBOARDING_SERVICE_TOKEN", ""),
		BackofficeURL:          toolkit.GetEnv("BACKOFFICE_INTERNAL_URL", ""),
		BackofficeServiceToken: toolkit.GetEnv("BACKOFFICE_INTERNAL_SERVICE_TOKEN", ""),
	}
	if k != nil {
		if cfg.Authn.JWTSecret == "" {
			cfg.Authn.JWTSecret = k.String("onboarding.auth.jwt_secret")
		}
		if cfg.Authn.JWTKey == "" {
			cfg.Authn.JWTKey = k.String("onboarding.auth.jwt_key")
		}
		if host := k.String("onboarding.iam.grpc_host"); host != "" {
			cfg.IAM.GRPCHost = host
		}
		if port := k.String("onboarding.iam.grpc_port"); port != "" {
			cfg.IAM.GRPCPort = port
		}
		if cfg.ServiceToken == "" {
			cfg.ServiceToken = k.String("onboarding.auth.service_token")
		}
		if cfg.BackofficeURL == "" {
			value := k.String("onboarding.backoffice.url")
			if value != "" && !strings.HasPrefix(value, "${") {
				cfg.BackofficeURL = value
			}
		}
		if cfg.BackofficeServiceToken == "" {
			cfg.BackofficeServiceToken = k.String("onboarding.backoffice.service_token")
			if strings.HasPrefix(cfg.BackofficeServiceToken, "${") {
				cfg.BackofficeServiceToken = ""
			}
		}
	}
	if cfg.BackofficeURL == "" {
		cfg.BackofficeURL = "http://localhost:8000"
	}
	return cfg
}
