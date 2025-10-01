package pdsql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFromViper_NoSub_ReturnsZeroConfig(t *testing.T) {
	v := viper.New()

	cfg, err := GetConfigFromViper("nonexistent", v)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Empty(t, cfg.URI)
	assert.Equal(t, ProviderType(""), cfg.Provider)
}

func TestGetConfigFromViper_WithSub_UnmarshalOK(t *testing.T) {
	v := viper.New()
	name := "main"
	base := "sql." + name

	v.Set(base+".uri", "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
	v.Set(base+".provider", "postgres")

	cfg, err := GetConfigFromViper(name, v)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "postgres://user:pass@localhost:5432/dbname?sslmode=disable", cfg.URI)
	assert.Equal(t, ProviderType("postgres"), cfg.Provider)
}

func TestNewDbFromConfig_Sqlmock_Succeeds(t *testing.T) {
	cfg := &Config{
		Provider: ProviderType("sqlmock"),
		URI:      "any",
	}

	db, err := NewDbFromConfig(cfg)
	require.NoError(t, err, "NewDbFromConfig should succeed for supported provider sqlmock")
	require.NotNil(t, db, "db should be non-nil on success")
	_ = db.Close()
}

func TestNewDbFromConfig_Sqlmock_Succeeds_WithDSN(t *testing.T) {
	const dsn = "pdsql-test-dsn"
	rawDB, _, err := sqlmock.NewWithDSN(dsn)
	require.NoError(t, err)
	defer rawDB.Close()

	cfg := &Config{
		Provider: ProviderType("sqlmock"),
		URI:      dsn,
	}

	db, err := NewDbFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	_ = db.Close()
}

func TestNewDbFromConfig_UnknownDriver_ReturnsConnectError(t *testing.T) {
	cfg := &Config{
		Provider: ProviderType("driver_that_does_not_exist"),
		URI:      "whatever://doesnot/matter",
	}
	db, err := NewDbFromConfig(cfg)
	require.Error(t, err)
	assert.Nil(t, db)
}
