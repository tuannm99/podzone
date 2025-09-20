package pdlog

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapFactory LoggerFactory = func(_ context.Context, cfg Config) (Logger, error) {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Env == "" {
		cfg.Env = "prod"
	}

	var zcfg zap.Config
	if cfg.Env == "prod" {
		zcfg = zap.NewProductionConfig()
		zcfg.Sampling = &zap.SamplingConfig{Initial: 100, Thereafter: 100}
	} else {
		zcfg = zap.NewDevelopmentConfig()
		zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
	}
	zcfg.Level = zap.NewAtomicLevelAt(lvl)
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	z, err := zcfg.Build()
	if err != nil {
		return nil, err
	}

	zl := &zapLogger{l: z}
	return zl.With("app", cfg.AppName, "env", cfg.Env), nil
}

type zapLogger struct{ l *zap.Logger }

func (z *zapLogger) With(kv ...any) Logger  { return &zapLogger{l: z.l.With(kvToZap(kv...)...)} }
func (z *zapLogger) Debug(msg string) Entry { return newZapEntry(z.l, LevelDebug, msg) }
func (z *zapLogger) Info(msg string) Entry  { return newZapEntry(z.l, LevelInfo, msg) }
func (z *zapLogger) Warn(msg string) Entry  { return newZapEntry(z.l, LevelWarn, msg) }
func (z *zapLogger) Error(msg string) Entry { return newZapEntry(z.l, LevelError, msg) }
func (z *zapLogger) Sync() error            { return z.l.Sync() }

type zapEntry struct {
	l     *zap.Logger
	level Level
	msg   string
	fs    []zap.Field
	err   error
}

func newZapEntry(l *zap.Logger, lvl Level, msg string) *zapEntry {
	return &zapEntry{l: l, level: lvl, msg: msg, fs: make([]zap.Field, 0, 8)}
}

func (e *zapEntry) With(kv ...any) Entry { e.fs = append(e.fs, kvToZap(kv...)...); return e }
func (e *zapEntry) Err(err error) Entry  { e.err = err; return e }
func (e *zapEntry) Send() {
	if e.err != nil {
		e.fs = append(e.fs, zap.NamedError("error", e.err))
	}
	switch e.level {
	case LevelDebug:
		e.l.Debug(e.msg, e.fs...)
	case LevelInfo:
		e.l.Info(e.msg, e.fs...)
	case LevelWarn:
		e.l.Warn(e.msg, e.fs...)
	default:
		e.l.Error(e.msg, e.fs...)
	}
}

func kvToZap(kv ...any) []zap.Field {
	if len(kv)%2 != 0 {
		kv = append(kv, "(MISSING)")
	}
	out := make([]zap.Field, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		key := toString(kv[i])
		val := kv[i+1]
		switch v := val.(type) {
		case string:
			out = append(out, zap.String(key, v))
		case int:
			out = append(out, zap.Int(key, v))
		case int64:
			out = append(out, zap.Int64(key, v))
		case bool:
			out = append(out, zap.Bool(key, v))
		case float64:
			out = append(out, zap.Float64(key, v))
		case time.Duration:
			out = append(out, zap.Duration(key, v))
		case error:
			out = append(out, zap.NamedError(key, v))
		default:
			out = append(out, zap.Any(key, v))
		}
	}
	return out
}

func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprint(s)
	}
}
