package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/polkiloo/go-musthave-metrics-tppl/migrations"

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
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

var open = func(ctx context.Context, dsn string) (Pool, error) {
	return pgxpool.New(ctx, dsn)
}

var migrate = func(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, ".")
}

func newDB(ctx context.Context, lc fx.Lifecycle, cfg *Config) (Pool, error) {
	if cfg == nil || cfg.DSN == "" {
		return nil, nil
	}

	if err := migrate(ctx, cfg.DSN); err != nil {
		return nil, err
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
