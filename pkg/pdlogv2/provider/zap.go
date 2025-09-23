package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZapFactory pdlogv2.FactoryFn = func(_ context.Context, cfg pdlogv2.Config) (pdlogv2.Logger, error) {
	levelText := cfg.Level
	if levelText == "" {
		levelText = "info"
	}
	env := cfg.Env
	if env == "" {
		env = "prod"
	}

	var zcfg zap.Config
	if env == "prod" {
		zcfg = zap.NewProductionConfig()
		zcfg.Sampling = &zap.SamplingConfig{Initial: 100, Thereafter: 100}
	} else {
		zcfg = zap.NewDevelopmentConfig()
		zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(levelText)); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", levelText)
	}
	zcfg.Level = zap.NewAtomicLevelAt(lvl)
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	z, err := zcfg.Build()
	if err != nil {
		return nil, err
	}

	zl := &zapLogger{l: z}
	return zl.With("app", cfg.AppName, "env", env), nil
}

type zapLogger struct{ l *zap.Logger }

func (z *zapLogger) With(kv ...any) pdlogv2.Logger { return &zapLogger{l: z.l.With(kvToZap(kv...)...)} }
func (z *zapLogger) Log(level pdlogv2.Level, msg string, kv ...any) {
	fs := kvToZap(kv...)
	switch level {
	case pdlogv2.LevelDebug:
		z.l.Debug(msg, fs...)
	case pdlogv2.LevelInfo:
		z.l.Info(msg, fs...)
	case pdlogv2.LevelWarn:
		z.l.Warn(msg, fs...)
	default:
		z.l.Error(msg, fs...)
	}
}
func (z *zapLogger) Debug(msg string, kv ...any) { z.Log(pdlogv2.LevelDebug, msg, kv...) }
func (z *zapLogger) Info(msg string, kv ...any)  { z.Log(pdlogv2.LevelInfo, msg, kv...) }
func (z *zapLogger) Warn(msg string, kv ...any)  { z.Log(pdlogv2.LevelWarn, msg, kv...) }
func (z *zapLogger) Error(msg string, kv ...any) { z.Log(pdlogv2.LevelError, msg, kv...) }
func (z *zapLogger) Sync() error                 { return z.l.Sync() }

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
