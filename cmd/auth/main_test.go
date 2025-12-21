package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	t.Parallel()

	app := newAppContainer()

	ctx := context.Background()
	require.NoError(t, app.Start(ctx))
	require.NoError(t, app.Stop(ctx))
}
