package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_App_Starts_And_Stops(t *testing.T) {
	t.Setenv("LOGGER_PROVIDER", "mock")
	t.Setenv("POSTGRES_AUTH_PROVIDER", "mock")
	t.Setenv("REDIS_AUTH_PROVIDER", "mock")

	t.Setenv("GRPC_PORT", "0")

	app := newAppContainer()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, app.Start(ctx), "app should start")
	require.NoError(t, app.Stop(ctx), "app should stop")
}

func Test_Main_DoesNotPanic(t *testing.T) {
	t.Setenv("LOGGER_PROVIDER", "mock")
	t.Setenv("POSTGRES_AUTH_PROVIDER", "mock")
	t.Setenv("REDIS_AUTH_PROVIDER", "mock")

	done := make(chan struct{})
	go func() {
		main()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}
}
