package pdsql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsurePostgresSchema_ValidatesInput(t *testing.T) {
	t.Run("missing admin DSN", func(t *testing.T) {
		err := EnsurePostgresSchema(context.Background(), "", "tenant_one")
		require.ErrorContains(t, err, "admin DSN is required")
	})

	t.Run("missing schema", func(t *testing.T) {
		err := EnsurePostgresSchema(context.Background(), "postgres://localhost/postgres", "")
		require.ErrorContains(t, err, "schema name is required")
	})
}
