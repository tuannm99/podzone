package pdsql

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/lib/pq"
)

var (
	sqlOpen     = sql.Open
	sqlxConnect = sqlx.Connect
)

func GetConfigFromKoanf(name string, k *koanf.Koanf) (*Config, error) {
	base := "sql." + name

	var cfg Config
	if err := k.Unmarshal(base, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", base, err)
	}
	return &cfg, nil
}

func NewDbFromConfig(cfg *Config) (*sqlx.DB, error) {
	switch cfg.Provider {
	case PostgresProvider:
		fullDSN := cfg.URI
		adminDSN, dbName, err := postgresAdminDSN(fullDSN)
		if err != nil {
			return nil, fmt.Errorf("parse postgres DSN: %w", err)
		}

		if err := ensurePostgresDatabase(adminDSN, dbName); err != nil {
			return nil, fmt.Errorf("ensure db %q: %w", dbName, err)
		}

		db, err := sqlxConnect("postgres", fullDSN)
		if err != nil {
			return nil, err
		}
		return db, nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

func postgresAdminDSN(dsn string) (adminDSN, dbName string, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", err
	}
	dbName = strings.TrimPrefix(u.Path, "/")
	u.Path = "/postgres"
	return u.String(), dbName, nil
}

func ensurePostgresDatabase(adminDSN, dbName string) error {
	admin, err := sqlOpen("postgres", adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close()

	var exists bool
	err = admin.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, dbName).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = admin.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(dbName))
	return err
}
