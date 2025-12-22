package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
)

// ConsulPublisher publishes runtime snapshot to Consul via KVStore.

type ConsulPublisher struct {
	kv kvstores.KVStore
}

func NewConsulPublisher(kv kvstores.KVStore) *ConsulPublisher {
	return &ConsulPublisher{kv: kv}
}

func (p *ConsulPublisher) Put(ctx context.Context, key string, value string) error {
	if err := p.kv.Put(key, []byte(value)); err != nil {
		return fmt.Errorf("consul put failed: %w", err)
	}
	return nil
}

func (p *ConsulPublisher) Delete(ctx context.Context, key string) error {
	if err := p.kv.Del(key); err != nil {
		return fmt.Errorf("consul delete failed: %w", err)
	}
	return nil
}

// DefaultBackoff returns a simple linear backoff for retries.
func DefaultBackoff(retry int) time.Duration {
	if retry <= 0 {
		return 2 * time.Second
	}
	if retry > 10 {
		retry = 10
	}
	return time.Duration(retry) * 5 * time.Second
}
