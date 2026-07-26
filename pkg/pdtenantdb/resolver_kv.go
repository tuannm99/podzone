package pdtenantdb

import (
	"context"
	"encoding/json"
	"errors"
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

// KVPlacementResolver resolves tenant placement from the configured KV store.
// Onboarding writes podzone/tenants/{tenantID}/placement when a postgres
// connection is registered. This resolver reads that key with a TTL cache.
type KVPlacementResolver struct {
	kv     kvstores.KVStore
	prefix string
	ttl    time.Duration
	// negTTL bounds how long a "no placement yet" result is cached, so a
	// tenant mid-onboarding doesn't hit the KV store on every request - see
	// MT9, docs/07-problems/multitenancy-2026-07-23.md. Deliberately
	// shorter than ttl (capped low) since a real placement showing up
	// should become visible fast.
	negTTL time.Duration

	mu       sync.RWMutex
	cache    map[string]cachedPlacement
	notFound map[string]time.Time
	sf       singleflight.Group
}

func NewKVPlacementResolver(kv kvstores.KVStore) PlacementResolver {
	return NewKVPlacementResolverWithTTL(kv, "podzone/tenants", 2*time.Minute)
}

func NewKVPlacementResolverWithTTL(kv kvstores.KVStore, prefix string, ttl time.Duration) PlacementResolver {
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	negTTL := ttl
	if negTTL <= 0 || negTTL > 15*time.Second {
		negTTL = 15 * time.Second
	}
	return &KVPlacementResolver{
		kv:       kv,
		prefix:   prefix,
		ttl:      ttl,
		negTTL:   negTTL,
		cache:    make(map[string]cachedPlacement),
		notFound: make(map[string]time.Time),
	}
}

func (r *KVPlacementResolver) Resolve(ctx context.Context, tenantID string) (Placement, error) {
	r.mu.RLock()
	if it, ok := r.cache[tenantID]; ok && time.Since(it.loaded) < r.ttl {
		r.mu.RUnlock()
		return it.pl, nil
	}
	if at, ok := r.notFound[tenantID]; ok && time.Since(at) < r.negTTL {
		r.mu.RUnlock()
		return Placement{}, fmt.Errorf("%w: tenant %s (cached)", ErrPlacementNotFound, tenantID)
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
		raw, err := r.kv.Get(ctx, key)
		if err != nil {
			if errors.Is(err, kvstores.ErrKeyNotFound) {
				r.mu.Lock()
				r.notFound[tenantID] = time.Now()
				r.mu.Unlock()
				return Placement{}, fmt.Errorf("%w: kv store get %s: %v", ErrPlacementNotFound, key, err)
			}
			return Placement{}, fmt.Errorf("%w: kv store get %s: %v", ErrPlacementBackend, key, err)
		}
		r.mu.Lock()
		delete(r.notFound, tenantID)
		r.mu.Unlock()

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
		if mode == ModeSchema && p.SchemaName == "" {
			return Placement{}, fmt.Errorf(
				"pdtenantdb: incomplete placement for tenant %s (missing schema_name for schema mode)",
				tenantID,
			)
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
