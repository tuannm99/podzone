package pdlog

import "fmt"


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

type (
	Level int
	// Config is usually loaded by app (e.g. via Viper) and injected.
	Config struct {
		Provider string `mapstructure:"provider"` // "zap" | "slog" | "noop"
		Level    string `mapstructure:"level"`    // "debug" | "info" | "warn" | "error"
		Env      string `mapstructure:"env"`      // "dev" | "prod"
		AppName  string `mapstructure:"app_name"` // set by caller
	}
)

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}
