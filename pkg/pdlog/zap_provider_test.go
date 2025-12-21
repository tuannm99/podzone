package pdlog

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestZapCore() (*zap.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		zapcore.AddSync(buf),
		zap.DebugLevel,
	)
	return zap.New(core), buf
}

func TestNewZapLogger_DefaultsAndFields(t *testing.T) {
	l, err := NewZapLogger(Config{
		AppName: "test-app",
	})
	require.NoError(t, err)
	require.NotNil(t, l)

	logWith := l.With("key", "value")
	require.NotNil(t, logWith)
}

func TestNewZapLogger_InvalidLevel(t *testing.T) {
	_, err := NewZapLogger(Config{
		Level: "invalid-level",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid log level")
}

func TestZapLogger_LogMethods(t *testing.T) {
	z, buf := newTestZapCore()
	logger := &zapLogger{l: z}

	logger.Debug("debug-msg", "k1", "v1")
	logger.Info("info-msg", "k2", 123)
	logger.Warn("warn-msg", "k3", true)
	logger.Error("error-msg", "k4", errors.New("boom"))

	_ = logger.Sync()

	out := buf.String()
	require.Contains(t, out, `"msg":"debug-msg"`)
	require.Contains(t, out, `"msg":"info-msg"`)
	require.Contains(t, out, `"msg":"warn-msg"`)
	require.Contains(t, out, `"msg":"error-msg"`)
	require.Contains(t, out, `"k4":"boom"`)
}

func TestKvToZap_AllTypes(t *testing.T) {
	errVal := errors.New("oops")
	fields := kvToZap(
		"string", "val",
		"int", 42,
		"int64", int64(64),
		"bool", true,
		"float", 3.14,
		"duration", time.Second,
		"err", errVal,
		"obj", map[string]int{"a": 1},
	)

	keys := make([]string, 0, len(fields))
	for _, f := range fields {
		keys = append(keys, f.Key)
	}
	require.Contains(t, keys, "string")
	require.Contains(t, keys, "err")
	require.Contains(t, keys, "obj")
}

func TestKvToZap_MissingValueAddsPlaceholder(t *testing.T) {
	fields := kvToZap("key-only")
	require.Equal(t, "key-only", fields[0].Key)
}
