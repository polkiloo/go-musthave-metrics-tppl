package db

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewDB_NilConfig(t *testing.T) {
	lc := fxtest.NewLifecycle(t)

	db, err := newDB(context.TODO(), lc, nil)
	if err == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db != nil {
		t.Fatalf("want nil db, got %v", db)
	}
}

func TestNewDB_OpenAndClose(t *testing.T) {
	lc := fxtest.NewLifecycle(t)
	cfg := &Config{DSN: "postgres://user@localhost/db"}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db, err := newDB(ctx, lc, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db == nil {
		t.Fatalf("db not created")
	}

	if err := lc.Start(ctx); err != nil {
		t.Fatalf("lc.Start: %v", err)
	}
	if err := lc.Stop(ctx); err != nil {
		t.Fatalf("lc.Stop: %v", err)
	}
}

func TestModule_ProvidesDB(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	openOrig := open
	open = func(ctx context.Context, dsn string) (Pool, error) { return mockPool, nil }
	defer func() { open = openOrig }()

	mockPool.ExpectClose()
	var got Pool
	app := fx.New(
		fx.Provide(func() context.Context { return context.TODO() }),
		fx.NopLogger,
		Module,
		fx.Supply(&Config{DSN: "postgres://user@localhost/db"}),
		fx.Invoke(func(p Pool) { got = p }),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	if got == nil {
		t.Fatalf("pool not provided")
	}
	if err := app.Stop(ctx); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
