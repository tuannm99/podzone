package pdredis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx/fxtest"
)

var log = &pdlog.NopLogger{}

func setupRedisContainer(t *testing.T) (testcontainers.Container, string) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	endpoint, err := redisC.Endpoint(ctx, "")
	require.NoError(t, err)

	return redisC, fmt.Sprintf("redis://%s/0", endpoint)
}

func TestNewClientFromConfig_Integration(t *testing.T) {
	ctx := context.Background()
	container, uri := setupRedisContainer(t)
	defer container.Terminate(ctx)

	cfg := &Config{URI: uri}
	client, err := NewClientFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	defer client.Close()

	err = client.Ping(ctx).Err()
	require.NoError(t, err)
}

func TestRegisterLifecycle_Integration(t *testing.T) {
	ctx := context.Background()
	container, uri := setupRedisContainer(t)
	defer container.Terminate(ctx)

	cfg := &Config{URI: uri}
	client, err := NewClientFromConfig(cfg)
	require.NoError(t, err)

	lc := fxtest.NewLifecycle(t)

	registerLifecycle(lc, client, log, cfg)

	startCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	require.NoError(t, lc.Start(startCtx))

	stopCtx, cancelStop := context.WithTimeout(ctx, 5*time.Second)
	defer cancelStop()
	require.NoError(t, lc.Stop(stopCtx))
}

func TestNewClientFromConfig_InvalidHost(t *testing.T) {
	cfg := &Config{URI: "redis://invalid:6379/0"}
	client, err := NewClientFromConfig(cfg)
	require.NoError(t, err) // redis.NewClient doesn't error on create

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = client.Ping(ctx).Err()
	require.Error(t, err)
}
