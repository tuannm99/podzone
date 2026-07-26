package routing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/backoffice/runtime/scope"
)

// PZEP-0018 / MT5: RequiredStoreScope must treat scope.CurrentStoreID as
// authoritative - an explicit value is only accepted when it agrees with
// the scope IAM already checked permission against, never as a silent
// override.
func TestRequiredStoreScope(t *testing.T) {
	withScope := func(storeID string) context.Context {
		return scope.WithStoreContext(context.Background(), scope.StoreContext{StoreID: storeID})
	}

	t.Run("empty explicit falls back to scoped value", func(t *testing.T) {
		got, err := RequiredStoreScope(withScope("store-1"), "")
		require.NoError(t, err)
		require.Equal(t, "store-1", got)
	})

	t.Run("explicit matching scoped value is accepted", func(t *testing.T) {
		got, err := RequiredStoreScope(withScope("store-1"), "store-1")
		require.NoError(t, err)
		require.Equal(t, "store-1", got)
	})

	t.Run("explicit differing from scoped value is rejected", func(t *testing.T) {
		_, err := RequiredStoreScope(withScope("store-1"), "store-2")
		require.ErrorIs(t, err, ErrStoreScopeMismatch)
	})

	t.Run("both empty is rejected", func(t *testing.T) {
		_, err := RequiredStoreScope(context.Background(), "")
		require.ErrorIs(t, err, ErrStoreScopeRequired)
	})

	t.Run("explicit value used when no scope is set", func(t *testing.T) {
		got, err := RequiredStoreScope(context.Background(), "store-1")
		require.NoError(t, err)
		require.Equal(t, "store-1", got)
	})
}
