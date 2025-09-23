package provider

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/tuannm99/podzone/pkg/pdlogv2"
)

var SlogFactory pdlogv2.FactoryFn = func(_ context.Context, cfg pdlogv2.Config) (pdlogv2.Logger, error) {
	lvl := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(cfg.Level)) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}

	var w io.Writer = os.Stdout
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl})
	l := slog.New(h)
	sl := &slogLogger{l: l}
	return sl.With("app", cfg.AppName, "env", cfg.Env), nil
}

type slogLogger struct{ l *slog.Logger }

func (s *slogLogger) With(kv ...any) pdlogv2.Logger { return &slogLogger{l: s.l.With(kv...)} }
func (s *slogLogger) Log(level pdlogv2.Level, msg string, kv ...any) {
	attrs := kvToSlog(kv...)   // []slog.Attr
	args := attrsToArgs(attrs) // []any
	switch level {
	case pdlogv2.LevelDebug:
		s.l.Debug(msg, args...)
	case pdlogv2.LevelInfo:
		s.l.Info(msg, args...)
	case pdlogv2.LevelWarn:
		s.l.Warn(msg, args...)
	default:
		s.l.Error(msg, args...)
	}
}
func (s *slogLogger) Debug(msg string, kv ...any) { s.Log(pdlogv2.LevelDebug, msg, kv...) }
func (s *slogLogger) Info(msg string, kv ...any)  { s.Log(pdlogv2.LevelInfo, msg, kv...) }
func (s *slogLogger) Warn(msg string, kv ...any)  { s.Log(pdlogv2.LevelWarn, msg, kv...) }
func (s *slogLogger) Error(msg string, kv ...any) { s.Log(pdlogv2.LevelError, msg, kv...) }
func (s *slogLogger) Sync() error                 { return nil }

// helpers
func kvToSlog(kv ...any) []slog.Attr {
	if len(kv)%2 != 0 {
		kv = append(kv, "(MISSING)")
	}
	out := make([]slog.Attr, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		key := toString(kv[i])
		val := kv[i+1]
		out = append(out, slog.Any(key, val))
	}
	return out
}

func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, 0, len(attrs))
	for _, a := range attrs {
		args = append(args, a)
	}
	return args
}
