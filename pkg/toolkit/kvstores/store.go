package kvstores

import (
	"context"
	"errors"
)

var ErrKeyNotFound = errors.New("kv key not found")

type KVStore interface {
	Get(ctx context.Context, path string) ([]byte, error)
	GetKVs(ctx context.Context, prefix string) (map[string][]byte, error)
	Put(ctx context.Context, path string, value []byte) error
	Del(ctx context.Context, path string) error
}
