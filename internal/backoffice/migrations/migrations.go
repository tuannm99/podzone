package migrations

import (
	"context"
	"embed"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
)

//go:embed sql/*.sql
var migrationsFS embed.FS

type migration struct {
	version string
	path    string
}

var appliedScopes sync.Map

func ApplyTx(ctx context.Context, tx *sqlx.Tx) error {
	scopeKey, err := migrationScopeKey(ctx, tx)
	if err != nil {
		return fmt.Errorf("resolve migration scope: %w", err)
	}
	if scopeKey != "" {
		if _, ok := appliedScopes.Load(scopeKey); ok {
			return nil
		}
		if _, err := tx.ExecContext(
			ctx,
			`SELECT pg_advisory_xact_lock($1)`,
			migrationLockKey(scopeKey),
		); err != nil {
			return fmt.Errorf("acquire migration lock: %w", err)
		}
		if _, ok := appliedScopes.Load(scopeKey); ok {
			return nil
		}
	}

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

	if scopeKey != "" {
		appliedScopes.Store(scopeKey, struct{}{})
	}
	return nil
}

func migrationScopeKey(ctx context.Context, tx *sqlx.Tx) (string, error) {
	var scope struct {
		Database string `db:"database_name"`
		Schema   string `db:"schema_name"`
	}
	if err := tx.GetContext(ctx, &scope, `
SELECT current_database() AS database_name, current_schema() AS schema_name
`); err != nil {
		return "", err
	}
	scope.Database = strings.TrimSpace(scope.Database)
	scope.Schema = strings.TrimSpace(scope.Schema)
	if scope.Database == "" {
		return "", nil
	}
	return scope.Database + "|" + scope.Schema, nil
}

func migrationLockKey(scopeKey string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(scopeKey))
	return int64(hasher.Sum64())
}
