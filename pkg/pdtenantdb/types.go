package pdtenantdb

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
