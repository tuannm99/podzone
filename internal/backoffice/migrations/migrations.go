package migrations

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

//go:embed sql/*.sql
var migrationsFS embed.FS

type migration struct {
	version string
	path    string
}

func ApplyTx(ctx context.Context, tx *sqlx.Tx) error {
	if _, err := tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS backoffice_schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("read migration dir: %w", err)
	}

	items := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version := strings.TrimSuffix(entry.Name(), ".sql")
		items = append(items, migration{
			version: version,
			path:    "sql/" + entry.Name(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].version < items[j].version
	})

	for _, item := range items {
		var exists bool
		if err := tx.GetContext(
			ctx,
			&exists,
			`SELECT EXISTS (SELECT 1 FROM backoffice_schema_migrations WHERE version = $1)`,
			item.version,
		); err != nil {
			return fmt.Errorf("check migration %s: %w", item.version, err)
		}
		if exists {
			continue
		}

		content, err := migrationsFS.ReadFile(item.path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", item.version, err)
		}
		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", item.version, err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO backoffice_schema_migrations (version) VALUES ($1)`,
			item.version,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", item.version, err)
		}
	}

	return nil
}
