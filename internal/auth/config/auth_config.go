package config

import "os"

type AuthConfig struct {
	JWTSecret      string
	JWTKey         string
	AppRedirectURL string
}

func NewAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTKey:         os.Getenv("JWT_KEY"),
		AppRedirectURL: os.Getenv("APP_REDIRECT_URL"),
	}
}
