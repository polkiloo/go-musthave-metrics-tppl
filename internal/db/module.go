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
	Begin(context.Context) (pgx.Tx, error)
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

type migrated struct{}

func runMigrations(ctx context.Context, cfg *Config) (migrated, error) {
	if cfg == nil || cfg.DSN == "" {
		return migrated{}, nil
	}
	return migrated{}, migrate(ctx, cfg.DSN)
}

func newPool(ctx context.Context, _ migrated, cfg *Config) (Pool, error) {
	if cfg == nil || cfg.DSN == "" {
		return nil, nil
	}
	return open(ctx, cfg.DSN)
}

func closePool(lc fx.Lifecycle, pool Pool) {
	if pool == nil {
		return
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
}

var Module = fx.Module(
	"db",
	fx.Provide(runMigrations, newPool),
	fx.Invoke(closePool),
)
