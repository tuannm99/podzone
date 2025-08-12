package config

import "os"

type AuthConfig struct {
	JWTSecret      []byte
	AppRedirectURL string
}

func NewAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:      []byte(os.Getenv("JWT_SECRET")),
		AppRedirectURL: os.Getenv("APP_REDIRECT_URL"),
	}
}
