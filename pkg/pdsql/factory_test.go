package pdsql

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFromViper_NoSub(t *testing.T) {
	v := viper.New()
	cfg, err := GetConfigFromViper("missing", v)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Empty(t, cfg.URI)
	assert.Empty(t, cfg.Provider)
}

func TestGetConfigFromViper_WithSub(t *testing.T) {
	v := viper.New()
	v.Set("sql.main.uri", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	v.Set("sql.main.provider", "postgres")

	cfg, err := GetConfigFromViper("main", v)
	require.NoError(t, err)
	assert.Equal(t, "postgres://user:pass@localhost:5432/db?sslmode=disable", cfg.URI)
	assert.Equal(t, ProviderType("postgres"), cfg.Provider)
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
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// mock sql.Open
	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}
	defer func() { sqlOpen = sql.Open }()

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("testdb").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	err = ensurePostgresDatabase("mockdsn", "testdb")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsurePostgresDatabase_CreateNew(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}
	defer func() { sqlOpen = sql.Open }()

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("newdb").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`CREATE DATABASE "newdb"`).WillReturnResult(sqlmock.NewResult(1, 1))

	err = ensurePostgresDatabase("mockdsn", "newdb")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsurePostgresDatabase_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}
	defer func() { sqlOpen = sql.Open }()

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("bad").
		WillReturnError(errors.New("query fail"))

	err = ensurePostgresDatabase("mockdsn", "bad")
	assert.ErrorContains(t, err, "query fail")
}

func TestEnsurePostgresDatabase_OpenError(t *testing.T) {
	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, errors.New("cannot open")
	}
	defer func() { sqlOpen = sql.Open }()

	err := ensurePostgresDatabase("mockdsn", "db")
	assert.ErrorContains(t, err, "cannot open")
}

func TestNewDbFromConfig_Postgres_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}
	sqlxConnect = func(driver, dsn string) (*sqlx.DB, error) {
		return sqlx.NewDb(db, "sqlmock"), nil
	}
	defer func() {
		sqlOpen = sql.Open
		sqlxConnect = sqlx.Connect
	}()

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("testdb").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	cfg := &Config{
		Provider: "postgres",
		URI:      "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
	}
	dbx, err := NewDbFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, dbx)
}

func TestNewDbFromConfig_Postgres_EnsureError(t *testing.T) {
	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, errors.New("open fail")
	}
	defer func() { sqlOpen = sql.Open }()

	cfg := &Config{Provider: "postgres", URI: "postgres://x:x@x/x"}
	dbx, err := NewDbFromConfig(cfg)
	require.Error(t, err)
	require.Nil(t, dbx)
}

func TestNewDbFromConfig_Unsupported(t *testing.T) {
	cfg := &Config{Provider: "mysql", URI: "mysql://localhost"}
	dbx, err := NewDbFromConfig(cfg)
	require.ErrorContains(t, err, "unsupported provider")
	require.Nil(t, dbx)
}
