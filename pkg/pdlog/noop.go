package pdlog

import "context"

type noopBackend struct{}

func (noopBackend) Name() string                                   { return "noop" }
func (noopBackend) New(context.Context, ...Option) (Logger, error) { return &noopLogger{}, nil }

type noopLogger struct{}

func (*noopLogger) With(...any) Logger { return &noopLogger{} }
func (*noopLogger) Debug(string) Entry { return &noopEntry{} }
func (*noopLogger) Info(string) Entry  { return &noopEntry{} }
func (*noopLogger) Warn(string) Entry  { return &noopEntry{} }
func (*noopLogger) Error(string) Entry { return &noopEntry{} }
func (*noopLogger) Sync() error        { return nil }

type noopEntry struct{}

func (*noopEntry) With(...any) Entry { return &noopEntry{} }
func (*noopEntry) Err(error) Entry   { return &noopEntry{} }
func (*noopEntry) Send()             {}
