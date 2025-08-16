package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ZapLogger struct {
	*zap.SugaredLogger
}

func (l *ZapLogger) WriteInfo(msg string, kv ...any) {
	if len(kv) > 0 {
		l.SugaredLogger.Infow(msg, kv...)
	} else {
		l.SugaredLogger.Info(msg)
	}
}
func (l *ZapLogger) WriteError(msg string, kv ...any) {
	if len(kv) > 0 {
		l.SugaredLogger.Errorw(msg, kv...)
	} else {
		l.SugaredLogger.Error(msg)
	}
}
func (l *ZapLogger) Sync() error {
	_ = l.SugaredLogger.Sync()
	return nil
}

var buildZapLogger = func(cfg zap.Config) (*zap.Logger, error) {
	return cfg.Build()
}

func New() (Logger, error) {
	cfg := zap.NewProductionConfig()

	z, err := buildZapLogger(cfg)
	if err != nil {
		return nil, err
	}

	return &ZapLogger{z.Sugar()}, nil
}

var Module = fx.Module(
	"zaplog",
	fx.Provide(New),
)
