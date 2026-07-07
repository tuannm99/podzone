package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
)

// KVStorePublisher publishes runtime routing projections.
type KVStorePublisher struct {
	kv kvstores.KVStore
}

func NewKVStorePublisher(kv kvstores.KVStore) *KVStorePublisher {
	return &KVStorePublisher{kv: kv}
}

func (p *KVStorePublisher) Put(ctx context.Context, key string, value string) error {
	if err := p.kv.Put(ctx, key, []byte(value)); err != nil {
		return fmt.Errorf("kv store put: %w", err)
	}
	return nil
}

func (p *KVStorePublisher) Delete(ctx context.Context, key string) error {
	if err := p.kv.Del(ctx, key); err != nil {
		return fmt.Errorf("kv store delete: %w", err)
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
