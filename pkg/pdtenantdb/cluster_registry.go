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

type ClusterConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

type ClusterRegistry interface {
	// GetCluster loads cluster config by cluster name (from Consul, cached).
	GetCluster(ctx context.Context, clusterName string) (ClusterConfig, error)
}

type cachedCluster struct {
	cfg    ClusterConfig
	loaded time.Time
}

type ConsulClusterRegistry struct {
	kv     kvstores.KVStore
	prefix string
	ttl    time.Duration

	mu    sync.RWMutex
	cache map[string]cachedCluster
	sf    singleflight.Group
}

func NewDefaultConsulClusterRegistry(kv kvstores.KVStore) ClusterRegistry {
	return NewConsulClusterRegistry(kv, "podzone/postgres/clusters", 2*time.Minute)
}

func NewConsulClusterRegistry(kv kvstores.KVStore, prefix string, ttl time.Duration) *ConsulClusterRegistry {
	if prefix == "" {
		prefix = "podzone/postgres/clusters"
	}
	if ttl == 0 {
		ttl = 2 * time.Minute
	}

	return &ConsulClusterRegistry{
		kv:     kv,
		prefix: prefix,
		ttl:    ttl,
		cache:  make(map[string]cachedCluster),
	}
}

func (r *ConsulClusterRegistry) GetCluster(ctx context.Context, clusterName string) (ClusterConfig, error) {
	// Fast path from cache
	r.mu.RLock()
	if it, ok := r.cache[clusterName]; ok && time.Since(it.loaded) < r.ttl {
		r.mu.RUnlock()
		return it.cfg, nil
	}
	r.mu.RUnlock()

	// Avoid thundering herd
	v, err, _ := r.sf.Do(clusterName, func() (any, error) {
		// Re-check cache
		r.mu.RLock()
		if it, ok := r.cache[clusterName]; ok && time.Since(it.loaded) < r.ttl {
			r.mu.RUnlock()
			return it.cfg, nil
		}
		r.mu.RUnlock()

		key := fmt.Sprintf("%s/%s", r.prefix, clusterName)
		raw, err := r.kv.Get(key)
		if err != nil {
			return ClusterConfig{}, err
		}

		var cfg ClusterConfig
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return ClusterConfig{}, fmt.Errorf("invalid cluster config json for %s: %w", key, err)
		}
		if cfg.Host == "" || cfg.Port == 0 {
			return ClusterConfig{}, fmt.Errorf("invalid cluster config (missing host/port) for %s", key)
		}

		r.mu.Lock()
		r.cache[clusterName] = cachedCluster{cfg: cfg, loaded: time.Now()}
		r.mu.Unlock()

		return cfg, nil
	})
	if err != nil {
		return ClusterConfig{}, err
	}
	return v.(ClusterConfig), nil
}
