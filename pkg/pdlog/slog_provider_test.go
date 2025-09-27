package pdlog

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestSlogLogger(buf *bytes.Buffer, level slog.Level) *slogLogger {
	h := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})
	return &slogLogger{l: slog.New(h)}
}

func TestNewSlogLogger_DefaultsAndFields(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	l := slog.New(h)

	sl := &slogLogger{l: l}
	logger := sl.With("app", "my-app", "env", "dev")

	logger.Info("hello", "key", "val")
	output := buf.String()

	require.Contains(t, output, `"msg":"hello"`)
	require.Contains(t, output, `"key":"val"`)
	require.Contains(t, output, `"app":"my-app"`)
}

func TestNewSlogLogger_LevelParsing(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestSlogLogger(buf, slog.LevelWarn)

	logger.Info("info-msg")
	logger.Warn("warn-msg")
	logger.Debug("debug-msg")
	logger.Error("err-msg")
	logger.Sync()

	out := buf.String()
	require.NotContains(t, out, `"msg":"info-msg"`)
	require.Contains(t, out, `"msg":"warn-msg"`)
}

func TestSlogLogger_LogMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestSlogLogger(buf, slog.LevelDebug)

	logger.Debug("debug-msg", "k1", "v1")
	logger.Info("info-msg", "k2", 42)
	logger.Warn("warn-msg", "k3", true)
	logger.Error("error-msg", "k4", "boom")

	output := buf.String()
	require.Contains(t, output, `"msg":"debug-msg"`)
	require.Contains(t, output, `"msg":"info-msg"`)
	require.Contains(t, output, `"msg":"warn-msg"`)
	require.Contains(t, output, `"msg":"error-msg"`)
}

func TestKvToSlog_AndAttrsToArgs(t *testing.T) {
	attrs := kvToSlog("key1", "val1", "key2", 42)
	require.Len(t, attrs, 2)
	require.Equal(t, "key1", attrs[0].Key)
	require.Equal(t, "val1", attrs[0].Value.String())

	// Odd number of kv should append (MISSING)
	attrs2 := kvToSlog("onlyKey")
	require.Len(t, attrs2, 1)
	require.Equal(t, "onlyKey", attrs2[0].Key)
	require.Equal(t, "(MISSING)", attrs2[0].Value.String())

	args := attrsToArgs(attrs2)
	require.Len(t, args, 1)
	require.IsType(t, slog.Attr{}, args[0])
}
