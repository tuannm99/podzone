package toolkit

import (
	"context"
	"sync"
)

type Factory[T any, C any] func(ctx context.Context, cfg C) (T, error)

type Registry[T any, C any] struct {
	mu        sync.RWMutex
	factories map[string]Factory[T, C]
	defaultID string
}

func NewRegistry[T any, C any](defaultID string) *Registry[T, C] {
	return &Registry[T, C]{
		factories: make(map[string]Factory[T, C]),
		defaultID: defaultID,
	}
}

func (r *Registry[T, C]) Register(id string, f Factory[T, C]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[id] = f
}

func (r *Registry[T, C]) Use(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.factories[id]; ok {
		r.defaultID = id
	}
}

func (r *Registry[T, C]) Get() Factory[T, C] {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.factories[r.defaultID]
}

func (r *Registry[T, C]) Lookup(id string) (Factory[T, C], bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[id]
	return f, ok
}
