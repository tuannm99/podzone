package config

import "github.com/tuannm99/podzone/pkg/toolkit"

type AuthConfig struct {
	JWTSecret      string
	JWTKey         string
	AppRedirectURL string
}

func NewAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:      toolkit.GetEnv("JWT_SECRET", ""),
		JWTKey:         toolkit.GetEnv("JWT_KEY", ""),
		AppRedirectURL: toolkit.GetEnv("APP_REDIRECT_URL", ""),
	}
}
