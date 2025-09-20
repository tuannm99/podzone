package pdlog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

var slogFactory LoggerFactory = func(_ context.Context, cfg Config) (Logger, error) {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	var lvl slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	var w io.Writer = os.Stdout // could be extended via cfg later
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl})
	l := slog.New(h)

	return (&slogLogger{l: l}).With("app", cfg.AppName, "env", cfg.Env), nil
}

type slogLogger struct{ l *slog.Logger }

func (s *slogLogger) With(kv ...any) Logger  { return &slogLogger{l: s.l.With(kv...)} }
func (s *slogLogger) Debug(msg string) Entry { return &slogEntry{l: s.l, lvl: LevelDebug, msg: msg} }
func (s *slogLogger) Info(msg string) Entry  { return &slogEntry{l: s.l, lvl: LevelInfo, msg: msg} }
func (s *slogLogger) Warn(msg string) Entry  { return &slogEntry{l: s.l, lvl: LevelWarn, msg: msg} }
func (s *slogLogger) Error(msg string) Entry { return &slogEntry{l: s.l, lvl: LevelError, msg: msg} }
func (s *slogLogger) Sync() error            { return nil }

type slogEntry struct {
	l   *slog.Logger
	lvl Level
	msg string
	kv  []any
	err error
}

func (e *slogEntry) With(kv ...any) Entry { e.kv = append(e.kv, kv...); return e }
func (e *slogEntry) Err(err error) Entry  { e.err = err; return e }
func (e *slogEntry) Send() {
	if e.err != nil {
		e.kv = append(e.kv, "error", e.err)
	}
	switch e.lvl {
	case LevelDebug:
		e.l.Debug(e.msg, e.kv...)
	case LevelInfo:
		e.l.Info(e.msg, e.kv...)
	case LevelWarn:
		e.l.Warn(e.msg, e.kv...)
	default:
		e.l.Error(e.msg, e.kv...)
	}
}
