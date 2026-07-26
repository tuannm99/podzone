package pdtenantdb

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

type Mode string

const (
	ModeSchema   Mode = "schema"   // schema-per-tenant in a shared database
	ModeDatabase Mode = "database" // database-per-tenant
)

type Config struct {
	SharedDB string // e.g. "backoffice"

	ConnMaxLifetime time.Duration
	MaxOpenConns    int
	MaxIdleConns    int

	// Dedicated DB pool controls (database-per-tenant)
	MaxDedicatedPools int
	DedicatedIdleTTL  time.Duration

	// OnSchemaBootstrap runs once per tenant scope the first time it is
	// resolved in this process: once per schema for schema-mode tenants
	// (right after CREATE SCHEMA IF NOT EXISTS succeeds), once per new
	// dedicated pool for database-mode tenants. schemaName is "" in the
	// database-mode case. The owning service uses this to apply its own
	// migrations without pkg/pdtenantdb importing anything service-specific,
	// and without a per-request round trip (see docs/07-problems/
	// multitenancy-2026-07-23.md MT8). Optional; nil is a no-op.
	OnSchemaBootstrap func(ctx context.Context, db *sqlx.DB, schemaName string) error

	// TenantRoleEnabled gates whether WithTenantTx issues `SET LOCAL ROLE`
	// into a per-tenant Postgres role (named after the tenant's own schema)
	// for schema-mode tenants. Off by default: false is byte-for-byte
	// today's (Phase 1) behavior. Requires the per-tenant role to already
	// exist for a given schema (see internal/backoffice/migrations/sql/
	// 0018_create_backoffice_tenant_role.sql) - if the role does not exist
	// yet, Postgres raises an error on SET LOCAL ROLE rather than silently
	// falling back, so only enable this after the role-provisioning
	// migration has run for every schema in use. See docs/09-pzep/
	// PZEP-0015-backoffice-per-tenant-db-role-phase-2.md (Phase 2 of MT1).
	TenantRoleEnabled bool

	// PlacementCacheTTL configures KVPlacementResolver's cache TTL when
	// wired via fxmodule.go. Zero uses the resolver's own default (2
	// minutes) - see MT9, docs/07-problems/multitenancy-2026-07-23.md.
	PlacementCacheTTL time.Duration

	// Logger receives structured log lines for pool lifecycle, placement
	// resolution failures, and dedicated-pool capacity rejections (MT16,
	// observability half only - see docs/09-pzep/
	// PZEP-0017-multitenancy-correctness-observability.md). Optional; a
	// nil Logger defaults to pdlog.NopLogger in NewManager.
	Logger pdlog.Logger

	// OnScopeInvalidated is called when WithTenantTx detects a schema it
	// believed was bootstrapped is actually gone (MT10, docs/07-problems/
	// multitenancy-2026-07-23.md). pdtenantdb evicts its own internal
	// bootstrapped cache itself; this callback lets the owning service
	// evict any cache it layered on top (e.g. internal/backoffice/
	// migrations.appliedScopes, which pdtenantdb cannot reach directly -
	// migrations imports pdtenantdb, not the other way around). Optional;
	// nil is a no-op.
	OnScopeInvalidated func(database, schemaName string)
}

type ConnKey struct {
	ClusterName string
	DBName      string
}

// Placement tells how to route a tenant.
type Placement struct {
	TenantID    string
	ClusterName string // <-- from onboarding (field: cluster_name)

	Mode       Mode
	DBName     string // schema-mode: shared DB, database-mode: bo_<tenant>
	SchemaName string // schema-mode: t_<tenant>
}

// PlacementResolver resolves tenant placement.
// You will implement this later using onboarding (+ cache).
type PlacementResolver interface {
	Resolve(ctx context.Context, tenantID string) (Placement, error)
}

var (
	ErrPlacementNotFound     = errors.New("pdtenantdb: placement not found")
	ErrPlacementBackend      = errors.New("pdtenantdb: placement backend error")
	ErrDedicatedPoolCapacity = errors.New("pdtenantdb: dedicated pool capacity reached")
)
