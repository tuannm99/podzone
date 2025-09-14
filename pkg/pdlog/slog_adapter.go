package pdlog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type slogBackend struct{}

func (slogBackend) Name() string { return "slog" }

func (slogBackend) New(_ context.Context, opts ...Option) (Logger, error) {
	o := &Options{Level: "info", Env: "dev", AppName: "app"}
	for _, f := range opts {
		f(o)
	}

	// writer có thể override qua Options.Extra["writer"]
	var w io.Writer = os.Stdout
	if v, ok := o.Extra["writer"].(io.Writer); ok && v != nil {
		w = v
	}

	// map level
	var lvl slog.Level
	switch strings.ToLower(o.Level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl})
	l := slog.New(h)

	return (&slogLogger{l: l}).With("app", o.AppName, "env", o.Env), nil
}

// ---- Logger adapter ----

type slogLogger struct {
	l *slog.Logger
}

func (s *slogLogger) With(kv ...any) Logger {
	return &slogLogger{l: s.l.With(kv...)}
}

func (s *slogLogger) Debug(msg string) Entry { return &slogEntry{l: s.l, lvl: LevelDebug, msg: msg} }
func (s *slogLogger) Info(msg string) Entry  { return &slogEntry{l: s.l, lvl: LevelInfo, msg: msg} }
func (s *slogLogger) Warn(msg string) Entry  { return &slogEntry{l: s.l, lvl: LevelWarn, msg: msg} }
func (s *slogLogger) Error(msg string) Entry { return &slogEntry{l: s.l, lvl: LevelError, msg: msg} }
func (s *slogLogger) Sync() error            { return nil } // slog no need flush

// ---- Entry builder ----

type slogEntry struct {
	l   *slog.Logger
	lvl Level
	msg string
	kv  []any
	err error
}

func (e *slogEntry) With(kv ...any) Entry {
	e.kv = append(e.kv, kv...)
	return e
}

func (e *slogEntry) Err(err error) Entry {
	e.err = err
	return e
}

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
