package backoffice

import (
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/migrations"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
)

// newTenantDBConfig wires backoffice.tenant_db (see internal/backoffice/config)
// into pdtenantdb.Config, including the migration-bootstrap hook. Previously
// this was `&pdtenantdb.Config{}` - unconfigurable, silently defaulting the
// connection pool regardless of environment. See docs/07-problems/
// multitenancy-2026-07-23.md MT2.
func newTenantDBConfig(cfg boconfig.Config, log pdlog.Logger) *pdtenantdb.Config {
	db := cfg.TenantDB
	return &pdtenantdb.Config{
		SharedDB:           db.SharedDB,
		ConnMaxLifetime:    db.ConnMaxLifetime(),
		MaxOpenConns:       db.MaxOpenConns,
		MaxIdleConns:       db.MaxIdleConns,
		MaxDedicatedPools:  db.MaxDedicatedPools,
		DedicatedIdleTTL:   db.DedicatedIdleTTL(),
		OnSchemaBootstrap:  migrations.ApplyToSchema,
		TenantRoleEnabled:  db.TenantRoleEnabled,
		PlacementCacheTTL:  db.PlacementCacheTTL(),
		Logger:             log,
		OnScopeInvalidated: migrations.InvalidateScope,
	}
}
