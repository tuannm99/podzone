package kvstores

import (
	"fmt"

	capi "github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

type KVStore interface {
	Get(path string) (string, error)
	Put(path string, value interface{}) error
	Del(path string) error
}

type ConsulKVStore struct {
	client *capi.Client
	logger *zap.Logger
}

func NewConsulKVStore(logger *zap.Logger) (*ConsulKVStore, error) {
	client, err := capi.NewClient(
		&capi.Config{
			Address: "localhost:8501",
			Token:   "8f3facc7-59ec-6b53-32e4-8590e99db44f",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error create consul client %s", err)
	}

	return &ConsulKVStore{client, logger}, nil
}

func (c *ConsulKVStore) Get(path string) (string, error) {
	kv := c.client.KV()

	pair, _, err := kv.Get(path, nil)
	if err != nil {
		return "", fmt.Errorf("error get kv %s", err)
	}
	fmt.Printf("KV: %v %s\n", pair.Key, pair.Value)

	return "", nil
}

func (c *ConsulKVStore) GetKeys(path string) (map[string]string, error) {
	kv := c.client.KV()

	// kv.Keys()
	pair, _, err := kv.Get(path, &capi.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get kv %s", err)
	}
	fmt.Printf("KV: %v %s\n", pair.Key, pair.Value)

	return nil, nil
}
