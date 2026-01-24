package pdserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type testLifecycle struct {
	hook fx.Hook
}

func (l *testLifecycle) Append(h fx.Hook) {
	l.hook = h
}

type logEntry struct {
	level pdlog.Level
	msg   string
	kv    []any
}

type testLogger struct {
	entries *[]logEntry
	fields  []any
}

func newTestLogger() *testLogger {
	entries := []logEntry{}
	return &testLogger{entries: &entries}
}

func (l *testLogger) With(kv ...any) pdlog.Logger {
	next := &testLogger{entries: l.entries}
	next.fields = append(append([]any{}, l.fields...), kv...)
	return next
}

func (l *testLogger) Log(level pdlog.Level, msg string, kv ...any) {
	l.entriesAppend(level, msg, kv...)
}

func (l *testLogger) Debug(msg string, kv ...any) { l.Log(pdlog.LevelDebug, msg, kv...) }
func (l *testLogger) Info(msg string, kv ...any)  { l.Log(pdlog.LevelInfo, msg, kv...) }
func (l *testLogger) Warn(msg string, kv ...any)  { l.Log(pdlog.LevelWarn, msg, kv...) }
func (l *testLogger) Error(msg string, kv ...any) { l.Log(pdlog.LevelError, msg, kv...) }
func (l *testLogger) Sync() error                 { return nil }

func (l *testLogger) entriesAppend(level pdlog.Level, msg string, kv ...any) {
	merged := append([]any{}, l.fields...)
	merged = append(merged, kv...)
	*l.entries = append(*l.entries, logEntry{level: level, msg: msg, kv: merged})
}

func findField(kv []any, key string) (string, bool) {
	for i := 0; i+1 < len(kv); i += 2 {
		if fmt.Sprint(kv[i]) == key {
			return fmt.Sprint(kv[i+1]), true
		}
	}
	return "", false
}

func TestRegisterHTTPServer_StartStop(t *testing.T) {
	lc := &testLifecycle{}
	log := newTestLogger()

	RegisterHTTPServer(
		lc,
		log,
		"127.0.0.1:0",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		WithComponent("unit"),
		WithStartMsg("start"),
		WithStopMsg("stop"),
		WithStoppedMsg("stopped"),
		WithServeErrorMsg("serve_err"),
		WithShutdownErrorMsg("shutdown_err"),
		WithLogFields("extra", "field"),
	)

	require.NotNil(t, lc.hook.OnStart)
	require.NoError(t, lc.hook.OnStart(context.Background()))

	entries := *log.entries
	require.NotEmpty(t, entries)
	require.Equal(t, "start", entries[0].msg)
	if v, ok := findField(entries[0].kv, "component"); !ok || v != "unit" {
		t.Fatalf("missing component field in log entry: %v", entries[0].kv)
	}
	if v, ok := findField(entries[0].kv, "extra"); !ok || v != "field" {
		t.Fatalf("missing extra field in log entry: %v", entries[0].kv)
	}

	require.NotNil(t, lc.hook.OnStop)
	require.NoError(t, lc.hook.OnStop(context.Background()))
}

func TestRegisterHTTPServer_BindError(t *testing.T) {
	lc := &testLifecycle{}
	log := newTestLogger()

	RegisterHTTPServer(
		lc,
		log,
		"127.0.0.1:notaport",
		http.NewServeMux(),
		WithComponent("unit"),
	)

	require.NotNil(t, lc.hook.OnStart)
	require.Error(t, lc.hook.OnStart(context.Background()))
}
