package logger

type Logger interface {
	WriteInfo(msg string, kv ...any)
	WriteError(msg string, kv ...any)
	Sync() error
}
