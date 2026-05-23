package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	os.Setenv("JWT_SECRET", "my_jwt")
	os.Setenv("JWT_KEY", "my_key")
	os.Setenv("APP_REDIRECT_URL", "redirect")
}

func TestNewAuthConfig(t *testing.T) {
	cfg := NewAuthConfig(nil)
	require.Equal(t, "my_jwt", cfg.JWTSecret)
	require.Equal(t, "my_key", cfg.JWTKey)
	require.Equal(t, "redirect", cfg.AppRedirectURL)
	require.Equal(t, "localhost", cfg.IAM.GRPCHost)
	require.Equal(t, "50053", cfg.IAM.GRPCPort)
}
