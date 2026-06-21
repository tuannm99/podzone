package pdsql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgresDSNWithDatabase(t *testing.T) {
	dsn, err := PostgresDSNWithDatabase(
		"postgres://postgres:secret@postgres:5432/postgres?sslmode=disable",
		"podzone_tenants",
	)

	require.NoError(t, err)
	require.Equal(t, "postgres://postgres:secret@postgres:5432/podzone_tenants?sslmode=disable", dsn)
}

func TestPostgresDSNWithDatabase_RequiresDatabase(t *testing.T) {
	_, err := PostgresDSNWithDatabase("postgres://postgres:secret@postgres:5432/postgres", "")

	require.ErrorContains(t, err, "database name is required")
}

func TestPostgresDSNWithDatabase_RequiresDSN(t *testing.T) {
	_, err := PostgresDSNWithDatabase("", "podzone_tenants")

	require.ErrorContains(t, err, "postgres DSN is required")
}
