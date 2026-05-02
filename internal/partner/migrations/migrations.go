package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var migrationsFS embed.FS

func Apply(ctx context.Context, db *sql.DB, dialect string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("goose.SetDialect: %w", err)
	}
	goose.SetBaseFS(migrationsFS)
	if err := goose.UpContext(ctx, db, "sql"); err != nil {
		return fmt.Errorf("goose.Up: %w", err)
	}
	return nil
}
