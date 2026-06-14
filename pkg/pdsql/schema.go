package pdsql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

func EnsurePostgresSchema(ctx context.Context, adminDSN string, schemaName string) error {
	adminDSN = strings.TrimSpace(adminDSN)
	if adminDSN == "" {
		return fmt.Errorf("postgres admin DSN is required")
	}
	schemaName = strings.TrimSpace(schemaName)
	if schemaName == "" {
		return fmt.Errorf("postgres schema name is required")
	}

	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres admin connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres admin connection: %w", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS `+pq.QuoteIdentifier(schemaName)); err != nil {
		return fmt.Errorf("create postgres schema: %w", err)
	}
	return nil
}
