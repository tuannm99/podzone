package pdlogv2

// Level represents log severity.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger is a simple, structured logger.
// - No Entry/Send chaining; just call Debug/Info/Warn/Error directly.
// - With(...) returns a derived logger with extra context (fields) bound. Preventing duplication
// Implementations must be safe for concurrent use.
type Logger interface {
	With(kv ...any) Logger
	Log(level Level, msg string, kv ...any)
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
	Sync() error
}
