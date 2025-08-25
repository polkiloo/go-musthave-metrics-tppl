package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	"go.uber.org/fx"
)

var (
	ErrMissConfig = fmt.Errorf("missing DB config")
)

type Config struct {
	DSN string
}

type Pool interface {
	Ping(context.Context) error
	Close()
}

var open = func(ctx context.Context, dsn string) (Pool, error) {
	return pgxpool.New(ctx, dsn)
}

func newDB(ctx context.Context, lc fx.Lifecycle, cfg *Config) (Pool, error) {
	if cfg == nil {
		return nil, ErrMissConfig
	}

	pool, err := open(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool, nil
}

var Module = fx.Module(
	"db",
	fx.Provide(newDB),
)
