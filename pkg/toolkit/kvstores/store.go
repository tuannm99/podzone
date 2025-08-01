package kvstores

type KVStore interface {
	Get(path string) ([]byte, error)
	GetKVs(prefix string) (map[string]string, error)
	Put(path string, value []byte) error
	Del(path string) error
}
