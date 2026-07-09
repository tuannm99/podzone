package provider

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/pdsql"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func (p *Provider) CheckDatabaseClusterHealth(
	ctx context.Context,
	cluster entity.DatabaseCluster,
) (entity.DatabaseClusterHealth, error) {
	checkedAt := time.Now().UTC()
	if strings.TrimSpace(p.cfg.AdminDSN) == "" {
		return entity.DatabaseClusterHealth{
			Healthy:   false,
			Message:   "postgres admin_dsn is required",
			CheckedAt: checkedAt,
		}, nil
	}
	placementDB := toolkit.FirstNonEmpty(cluster.PlacementDB, p.cfg.DBName, "podzone_tenants")
	dsn, err := pdsql.PostgresDSNWithDatabase(p.cfg.AdminDSN, placementDB)
	if err != nil {
		return entity.DatabaseClusterHealth{}, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return entity.DatabaseClusterHealth{}, fmt.Errorf("open postgres placement database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return entity.DatabaseClusterHealth{
			Healthy:   false,
			Message:   "postgres ping failed: " + err.Error(),
			CheckedAt: checkedAt,
		}, nil
	}

	schemas, err := countTenantSchemas(ctx, db, toolkit.FirstNonEmpty(p.cfg.SchemaPrefix, "t_"))
	if err != nil {
		return entity.DatabaseClusterHealth{}, err
	}
	connections, err := countDatabaseConnections(ctx, db, placementDB)
	if err != nil {
		return entity.DatabaseClusterHealth{}, err
	}
	return entity.DatabaseClusterHealth{
		Healthy:            true,
		CurrentSchemas:     schemas,
		CurrentConnections: connections,
		Message:            "postgres placement database is reachable",
		CheckedAt:          checkedAt,
	}, nil
}

func countTenantSchemas(ctx context.Context, db *sql.DB, schemaPrefix string) (int, error) {
	var count int
	err := db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM information_schema.schemata WHERE starts_with(schema_name, $1)`,
		schemaPrefix,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count tenant schemas: %w", err)
	}
	return count, nil
}

func countDatabaseConnections(ctx context.Context, db *sql.DB, dbName string) (int, error) {
	var count int
	err := db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`,
		dbName,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count database connections: %w", err)
	}
	return count, nil
}
