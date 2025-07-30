package kvstores

import (
	"fmt"

	capi "github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

type KVStore interface {
	Get(path string) ([]byte, error)
	GetKVs(prefix string) (map[string]string, error)
	Put(path string, value []byte) error
	Del(path string) error
}

type ConsulKVStore struct {
	client *capi.Client
	logger *zap.Logger
}

func NewConsulKVStore(logger *zap.Logger, address, token string) (*ConsulKVStore, error) {
	client, err := capi.NewClient(
		&capi.Config{
			Address: address,
			Token:   token,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create consul client %s", err)
	}

	return &ConsulKVStore{client, logger}, nil
}

func (c *ConsulKVStore) Get(path string) ([]byte, error) {
	kv := c.client.KV()

	pair, _, err := kv.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get kv %s", err)
	}

	c.logger.Debug("KV", zap.String("key", pair.Key), zap.ByteString("value", pair.Value))

	return pair.Value, nil
}

func (c *ConsulKVStore) GetKVs(prefix string) (map[string][]byte, error) {
	kv := c.client.KV()

	kvPairs, _, err := kv.List(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get keys %s", err)
	}

	result := make(map[string][]byte)
	keys := []string{}
	for _, kvPair := range kvPairs {
		result[kvPair.Key] = kvPair.Value
		keys = append(keys, kvPair.Key)
	}

	c.logger.Debug("KV", zap.Strings("keys", keys))

	return result, nil
}

func (c *ConsulKVStore) Put(path string, val []byte) error {
	kv := c.client.KV()

	_, err := kv.Put(&capi.KVPair{Key: path, Value: val}, nil)
	if err != nil {
		return fmt.Errorf("cannot put kv %s", err)
	}

	c.logger.Debug("put KV", zap.String("path", path), zap.ByteString("byte", val))

	return nil
}

func (c *ConsulKVStore) Del(path string) error {
	kv := c.client.KV()

	_, err := kv.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("cannot delete kv %s", err)
	}

	c.logger.Debug("KV deleted", zap.String("path", path))

	return nil
}
