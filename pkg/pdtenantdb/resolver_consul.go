package pdtenantdb

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
	"golang.org/x/sync/singleflight"
)

type placementJSON struct {
	ClusterName string `json:"cluster_name"`
	Mode        string `json:"mode"`
	DBName      string `json:"db_name"`
	SchemaName  string `json:"schema_name"`
}

type cachedPlacement struct {
	pl     Placement
	loaded time.Time
}

// ConsulPlacementResolver resolves tenant placement from Consul KV.
// Onboarding writes podzone/tenants/{tenantID}/placement when a postgres
// connection is registered. This resolver reads that key with a TTL cache.
type ConsulPlacementResolver struct {
	kv     kvstores.KVStore
	prefix string
	ttl    time.Duration

	mu    sync.RWMutex
	cache map[string]cachedPlacement
	sf    singleflight.Group
}

func NewConsulPlacementResolver(kv kvstores.KVStore) PlacementResolver {
	return NewConsulPlacementResolverWithTTL(kv, "podzone/tenants", 2*time.Minute)
}

func NewConsulPlacementResolverWithTTL(kv kvstores.KVStore, prefix string, ttl time.Duration) PlacementResolver {
	return &ConsulPlacementResolver{
		kv:     kv,
		prefix: prefix,
		ttl:    ttl,
		cache:  make(map[string]cachedPlacement),
	}
}

func (r *ConsulPlacementResolver) Resolve(ctx context.Context, tenantID string) (Placement, error) {
	r.mu.RLock()
	if it, ok := r.cache[tenantID]; ok && time.Since(it.loaded) < r.ttl {
		r.mu.RUnlock()
		return it.pl, nil
	}
	r.mu.RUnlock()

	v, err, _ := r.sf.Do(tenantID, func() (any, error) {
		// Re-check after acquiring the singleflight slot.
		r.mu.RLock()
		if it, ok := r.cache[tenantID]; ok && time.Since(it.loaded) < r.ttl {
			r.mu.RUnlock()
			return it.pl, nil
		}
		r.mu.RUnlock()

		key := fmt.Sprintf("%s/%s/placement", r.prefix, tenantID)
		raw, err := r.kv.Get(key)
		if err != nil {
			return Placement{}, fmt.Errorf("%w: consul get %s: %v", ErrPlacementNotFound, key, err)
		}

		var p placementJSON
		if err := json.Unmarshal(raw, &p); err != nil {
			return Placement{}, fmt.Errorf("pdtenantdb: invalid placement json for tenant %s: %w", tenantID, err)
		}
		if p.ClusterName == "" || p.DBName == "" {
			return Placement{}, fmt.Errorf(
				"pdtenantdb: incomplete placement for tenant %s (missing cluster_name or db_name)",
				tenantID,
			)
		}

		mode := ModeSchema
		if p.Mode == string(ModeDatabase) {
			mode = ModeDatabase
		}

		pl := Placement{
			TenantID:    tenantID,
			ClusterName: p.ClusterName,
			Mode:        mode,
			DBName:      p.DBName,
			SchemaName:  p.SchemaName,
		}

		r.mu.Lock()
		r.cache[tenantID] = cachedPlacement{pl: pl, loaded: time.Now()}
		r.mu.Unlock()

		return pl, nil
	})
	if err != nil {
		return Placement{}, err
	}
	return v.(Placement), nil
}
