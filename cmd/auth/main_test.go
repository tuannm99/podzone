package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	t.Parallel()

	app := newAppContainer()

	ctx := context.Background()
	err := app.Start(ctx)
	if err != nil && strings.Contains(err.Error(), "operation not permitted") {
		t.Skipf("skipping auth app start in restricted environment: %v", err)
	}
	require.NoError(t, err)
	require.NoError(t, app.Stop(ctx))
}
