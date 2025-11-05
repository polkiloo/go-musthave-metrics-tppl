package logger

// Logger abstracts structured logging used across the application.
type Logger interface {
	WriteInfo(msg string, kv ...any)
	WriteError(msg string, kv ...any)
	Sync() error
}
