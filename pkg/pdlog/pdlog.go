package pdlog

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger interface {
	With(kv ...any) Logger
	Debug(msg string) Entry
	Info(msg string) Entry
	Warn(msg string) Entry
	Error(msg string) Entry
	Sync() error
}

type Entry interface {
	With(kv ...any) Entry
	Err(err error) Entry
	Send()
}
