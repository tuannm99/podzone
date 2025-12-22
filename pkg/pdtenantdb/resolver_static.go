package pdtenantdb

import (
	"context"
	"fmt"
	"strings"
)

// StaticPlacementResolver is a simple resolver for dev/testing.
// It derives schema/db from tenantID and uses a default cluster_name.
type StaticPlacementResolver struct {
	// DefaultClusterName is used when there is no per-tenant override.
	DefaultClusterName string

	// SharedDB is the shared database name for schema-per-tenant mode.
	SharedDB string

	// BigTenants contains tenants that should use database-per-tenant mode.
	BigTenants map[string]bool
}

func NewStaticPlacementResolver(cfg *Config) PlacementResolver {
	// Minimal defaults; adjust as needed.
	return &StaticPlacementResolver{
		DefaultClusterName: "pg-default",
		SharedDB:           cfg.SharedDB,
		BigTenants:         map[string]bool{},
	}
}

func (r *StaticPlacementResolver) Resolve(ctx context.Context, tenantID string) (Placement, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return Placement{}, ErrPlacementNotFound
	}

	// Default cluster
	cluster := r.DefaultClusterName
	if cluster == "" {
		return Placement{}, fmt.Errorf("pdtenantdb: missing DefaultClusterName in resolver")
	}

	// Large tenant -> dedicated database
	if r.BigTenants != nil && r.BigTenants[tenantID] {
		return Placement{
			TenantID:    tenantID,
			ClusterName: cluster,
			Mode:        ModeDatabase,
			DBName:      "bo_" + sanitizeName(tenantID),
			// SchemaName optional (usually public)
		}, nil
	}

	// Default -> schema-per-tenant in shared database
	sharedDB := r.SharedDB
	if sharedDB == "" {
		sharedDB = "backoffice"
	}
	return Placement{
		TenantID:    tenantID,
		ClusterName: cluster,
		Mode:        ModeSchema,
		DBName:      sharedDB,
		SchemaName:  "t_" + sanitizeName(tenantID),
	}, nil
}

// sanitizeName is a minimal sanitizer for schema/db naming.
// NOTE: Keep it deterministic. For production, use a stronger sanitizer.
func sanitizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// Replace invalid chars with underscore.
	var b strings.Builder
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			b.WriteRune(ch)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		return "unknown"
	}
	return out
}
