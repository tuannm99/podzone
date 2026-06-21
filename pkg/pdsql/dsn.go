package pdsql

import (
	"fmt"
	"net/url"
	"strings"
)

func PostgresDSNWithDatabase(dsn string, dbName string) (string, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return "", fmt.Errorf("postgres DSN is required")
	}
	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return "", fmt.Errorf("postgres database name is required")
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	u.Path = "/" + dbName
	return u.String(), nil
}

func EnsurePostgresDatabase(adminDSN string, dbName string) error {
	adminDSN = strings.TrimSpace(adminDSN)
	if adminDSN == "" {
		return fmt.Errorf("postgres admin DSN is required")
	}
	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return fmt.Errorf("postgres database name is required")
	}
	return ensurePostgresDatabase(adminDSN, dbName)
}
