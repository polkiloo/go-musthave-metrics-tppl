package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestNewPool_NilConfig(t *testing.T) {
	m, err := runMigrations(context.TODO(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	db, err := newPool(context.TODO(), m, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db != nil {
		t.Fatalf("want nil db, got %v", db)
	}
}

func TestNewPool_EmptyDSN(t *testing.T) {
	m, err := runMigrations(context.TODO(), &Config{DSN: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	db, err := newPool(context.TODO(), m, &Config{DSN: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db != nil {
		t.Fatalf("want nil db, got %v", db)
	}
}

func TestNewPool_OpenAndClose(t *testing.T) {
	lc := fxtest.NewLifecycle(t)
	cfg := &Config{DSN: "postgres://user@localhost/db"}

	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	openOrig := open
	migrateOrig := migrate
	open = func(ctx context.Context, dsn string) (Pool, error) { return mockPool, nil }
	migrate = func(context.Context, string) error { return nil }
	defer func() { open = openOrig; migrate = migrateOrig }()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	m, err := runMigrations(ctx, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	db, err := newPool(ctx, m, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db == nil {
		t.Fatalf("db not created")
	}

	mockPool.ExpectClose()
	closePool(lc, db)

	if err := lc.Start(ctx); err != nil {
		t.Fatalf("lc.Start: %v", err)
	}
	if err := lc.Stop(ctx); err != nil {
		t.Fatalf("lc.Stop: %v", err)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestModule_ProvidesDB(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	openOrig := open
	migrateOrig := migrate
	open = func(ctx context.Context, dsn string) (Pool, error) { return mockPool, nil }
	migrate = func(context.Context, string) error { return nil }
	defer func() { open = openOrig; migrate = migrateOrig }()

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

func TestMigrate_WithTestcontainersPostgres(t *testing.T) {
	t.Skip()
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	const (
		dbName  = "appdb"
		dbUser  = "app"
		dbPass  = "secret"
		dbImage = "postgres:16-alpine"
	)
	ctr, err := postgres.Run(
		ctx,
		dbImage,
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, ctr)

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable", "application_name=migrate_test")
	require.NoError(t, err)

	require.NoError(t, migrate(ctx, dsn), "migrate should succeed against fresh container")

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer db.Close()

	var n int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT COUNT(*) FROM goose_db_version`).Scan(&n))
}
