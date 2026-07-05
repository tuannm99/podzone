package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServerConfigDefaultsToAuthGRPCPort(t *testing.T) {
	t.Setenv("AUTH_GRPC_HOST", "")
	t.Setenv("AUTH_GRPC_PORT", "")

	cfg := NewServerConfig(nil)

	require.Equal(t, "localhost", cfg.Auth.GRPCHost)
	require.Equal(t, "50051", cfg.Auth.GRPCPort)
}
