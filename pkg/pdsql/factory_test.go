package pdsql

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/testkit"
)

func TestGetConfigFromKoanf_NoSub(t *testing.T) {
	k := koanf.New(".")

	cfg, err := GetConfigFromKoanf("missing", k)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Empty(t, cfg.URI)
	assert.Empty(t, cfg.Provider)
	assert.False(t, cfg.ShouldRunMigration)
}

func TestGetConfigFromKoanf_WithSub(t *testing.T) {
	k := koanf.New(".")

	k.Set("sql.main.uri", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	k.Set("sql.main.provider", "postgres")
	k.Set("sql.main.should_run_migration", true)

	cfg, err := GetConfigFromKoanf("main", k)
	require.NoError(t, err)

	assert.Equal(t, "postgres://user:pass@localhost:5432/db?sslmode=disable", cfg.URI)
	assert.Equal(t, ProviderType("postgres"), cfg.Provider)
	assert.True(t, cfg.ShouldRunMigration)
}

func TestPostgresAdminDSN_OK(t *testing.T) {
	admin, dbName, err := postgresAdminDSN("postgres://user:pass@host:5432/mydb?sslmode=disable")
	require.NoError(t, err)
	assert.Contains(t, admin, "/postgres")
	assert.Equal(t, "mydb", dbName)
}

func TestPostgresAdminDSN_Invalid(t *testing.T) {
	_, _, err := postgresAdminDSN("://badurl")
	assert.Error(t, err)
}

func TestEnsurePostgresDatabase_Exists(t *testing.T) {
	dbName := fmt.Sprintf("podzone_exists_%d", time.Now().UnixNano())
	full := testkit.PostgresDSNWithDB(t, dbName)
	admin, _, err := postgresAdminDSN(full)
	require.NoError(t, err)

	require.NoError(t, ensurePostgresDatabase(admin, dbName))
	require.NoError(t, ensurePostgresDatabase(admin, dbName))
}

func TestEnsurePostgresDatabase_CreateNew(t *testing.T) {
	dbName := fmt.Sprintf("podzone_new_%d", time.Now().UnixNano())
	full := testkit.PostgresDSNWithDB(t, dbName)
	admin, _, err := postgresAdminDSN(full)
	require.NoError(t, err)

	require.NoError(t, ensurePostgresDatabase(admin, dbName))

	dbx, err := sqlx.Connect("postgres", full)
	require.NoError(t, err)
	_ = dbx.Close()
}

func TestNewDbFromConfig_Postgres_Success(t *testing.T) {
	dbName := fmt.Sprintf("podzone_cfg_%d", time.Now().UnixNano())
	full := testkit.PostgresDSNWithDB(t, dbName)

	cfg := &Config{
		Provider: "postgres",
		URI:      full,
	}
	dbx, err := NewDbFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, dbx)
	_ = dbx.Close()
}

func TestNewDbFromConfig_Unsupported(t *testing.T) {
	cfg := &Config{Provider: "mysql", URI: "mysql://localhost"}
	dbx, err := NewDbFromConfig(cfg)
	require.ErrorContains(t, err, "unsupported provider")
	require.Nil(t, dbx)
}
