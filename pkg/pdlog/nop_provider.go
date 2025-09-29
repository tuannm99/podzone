package pdlog

type NopLogger struct{}

func (NopLogger) With(...any) Logger        { return NopLogger{} }
func (NopLogger) Log(Level, string, ...any) {}
func (NopLogger) Debug(string, ...any)      {}
func (NopLogger) Info(string, ...any)       {}
func (NopLogger) Warn(string, ...any)       {}
func (NopLogger) Error(string, ...any)      {}
func (NopLogger) Sync() error               { return nil }
