package pdlog

type NopLogger struct{}

func (NopLogger) With(...any) Logger {
	return NopLogger{}
}

func (NopLogger) Log(Level, string, ...any) {
	// just for mock
}

func (NopLogger) Debug(string, ...any) {
	// just for mock
}

func (NopLogger) Info(string, ...any) {
	// just for mock
}

func (NopLogger) Warn(string, ...any) {
	// just for mock
}

func (NopLogger) Error(string, ...any) {
	// just for mock
}

func (NopLogger) Sync() error { return nil }
