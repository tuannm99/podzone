package provider

import (
	"context"
	"sync"

	"github.com/tuannm99/podzone/pkg/pdlogv2"
)

type Record struct {
	Level pdlogv2.Level
	Msg   string
	KVs   []any
}

type mockLogger struct {
	mu  sync.Mutex
	kvs []any    // bound from With(...)
	rec []Record // captured logs
}

func (m *mockLogger) With(kv ...any) pdlogv2.Logger {
	cp := &mockLogger{}
	cp.mu.Lock()
	cp.kvs = append(cp.kvs, m.kvs...)
	cp.kvs = append(cp.kvs, kv...)
	cp.mu.Unlock()
	return cp
}

func (m *mockLogger) Log(level pdlogv2.Level, msg string, kv ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := Record{Level: level, Msg: msg, KVs: append(append([]any{}, m.kvs...), kv...)}
	m.rec = append(m.rec, r)
}
func (m *mockLogger) Debug(msg string, kv ...any) { m.Log(pdlogv2.LevelDebug, msg, kv...) }
func (m *mockLogger) Info(msg string, kv ...any)  { m.Log(pdlogv2.LevelInfo, msg, kv...) }
func (m *mockLogger) Warn(msg string, kv ...any)  { m.Log(pdlogv2.LevelWarn, msg, kv...) }
func (m *mockLogger) Error(msg string, kv ...any) { m.Log(pdlogv2.LevelError, msg, kv...) }
func (m *mockLogger) Sync() error                 { return nil }

// helpers for tests (type assert to access)
func (m *mockLogger) Records() []Record {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Record, len(m.rec))
	copy(out, m.rec)
	return out
}

func (m *mockLogger) Reset() {
	m.mu.Lock()
	m.rec = nil
	m.mu.Unlock()
}

var MockFactory pdlogv2.FactoryFn = func(_ context.Context, _ pdlogv2.Config) (pdlogv2.Logger, error) {
	return &mockLogger{}, nil
}
