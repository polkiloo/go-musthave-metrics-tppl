package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ZapLogger is a Logger backed by Uber's zap.SugaredLogger.
type ZapLogger struct {
	*zap.SugaredLogger
}

// WriteInfo logs an informational message with optional key-value pairs.
func (l *ZapLogger) WriteInfo(msg string, kv ...any) {
	if len(kv) > 0 {
		l.SugaredLogger.Infow(msg, kv...)
	} else {
		l.SugaredLogger.Info(msg)
	}
}

// WriteError logs an error message with optional key-value pairs.
func (l *ZapLogger) WriteError(msg string, kv ...any) {
	if len(kv) > 0 {
		l.SugaredLogger.Errorw(msg, kv...)
	} else {
		l.SugaredLogger.Error(msg)
	}
}

// Sync flushes buffered log entries.
func (l *ZapLogger) Sync() error {
	_ = l.SugaredLogger.Sync()
	return nil
}

var buildZapLogger = func(cfg zap.Config) (*zap.Logger, error) {
	return cfg.Build()
}

// NewZapLogger constructs the production zap logger wrapped in the Logger interface.
func NewZapLogger() (Logger, error) {
	cfg := zap.NewProductionConfig()

	z, err := buildZapLogger(cfg)
	if err != nil {
		return nil, err
	}

	return &ZapLogger{z.Sugar()}, nil
}

// Module registers the zap logger within the fx container.
var Module = fx.Module(
	"zaplog",
	fx.Provide(NewZapLogger),
)
