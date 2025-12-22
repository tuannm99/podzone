package pdtenantdb

import (
	"context"
	"errors"
	"time"

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
	ErrDedicatedPoolCapacity = errors.New("pdtenantdb: dedicated pool capacity reached")
)
