package pdlog

import (
	"context"
	"sync"
)

func init() {
	Register(zapBackend{})
	Register(slogBackend{})
	Register(noopBackend{})
}

var (
	mu                 sync.RWMutex
	adapters           = map[string]BackendAdapter{}
	defaultLogProvider = "zap"
)

func Register(a BackendAdapter) {
	mu.Lock()
	defer mu.Unlock()
	adapters[a.Name()] = a
}

func Resolve(name string) (BackendAdapter, bool) {
	mu.RLock()
	defer mu.RUnlock()
	a, ok := adapters[name]
	return a, ok
}

func SetDefault(name string) { defaultLogProvider = name }

func NewFrom(name string, ctx context.Context, opts ...Option) (Logger, error) {
	mu.RLock()
	a, ok := adapters[name]
	if !ok {
		a = adapters[defaultLogProvider]
	}
	mu.RUnlock()
	return a.New(ctx, opts...)
}
