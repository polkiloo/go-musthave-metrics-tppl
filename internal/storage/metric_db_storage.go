package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/retrier"
)

const (
	sqlUpdateGauges = `INSERT INTO gauges(id, value, updated_at) VALUES ($1,$2,NOW())
    ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();`

	sqlUpdateCounters = `INSERT INTO counters(id, value, updated_at) VALUES ($1,$2,NOW())
    ON CONFLICT (id) DO UPDATE SET value = counters.value + EXCLUDED.value, updated_at = NOW();`

	sqlSetGauges = `INSERT INTO gauges(id, value, updated_at) VALUES ($1,$2,NOW())
    ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();`

	sqlSetCounters = `INSERT INTO counters(id, value, updated_at) VALUES ($1,$2,NOW())
    ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();`

	sqlUpsertGauges = `
	INSERT INTO gauges(id, value, updated_at)
	SELECT u.id, u.value, NOW()
	FROM UNNEST($1::text[], $2::double precision[]) AS u(id, value)
	ON CONFLICT (id) DO UPDATE
	SET value = EXCLUDED.value,
    updated_at = NOW();`

	sqlUpsertCounters = `
	INSERT INTO counters(id, value, updated_at)
	SELECT u.id, u.value, NOW()
	FROM UNNEST($1::text[], $2::bigint[]) AS u(id, value)
	ON CONFLICT (id) DO UPDATE
	SET value = counters.value + EXCLUDED.value,
	updated_at = NOW();`
)

// DBStorage persists metrics in PostgreSQL.
type DBStorage struct {
	pool db.Pool
}

// NewDBStorage constructs a database-backed MetricStorage implementation.
func NewDBStorage(p db.Pool) *DBStorage {
	return &DBStorage{pool: p}
}

// UpdateGauge upserts a gauge metric value.
func (s *DBStorage) UpdateGauge(name string, value float64) {
	_ = retrier.Do(context.Background(), func() error {
		_, err := s.pool.Exec(context.Background(), sqlUpdateGauges,
			name, value,
		)
		return err
	}, isPGConnError, retrier.DefaultDelays)
}

// UpdateCounter increments a counter metric in the database.
func (s *DBStorage) UpdateCounter(name string, delta int64) {
	_ = retrier.Do(context.Background(), func() error {
		_, err := s.pool.Exec(context.Background(), sqlUpdateCounters,
			name, delta,
		)
		return err
	}, isPGConnError, retrier.DefaultDelays)
}

// GetGauge retrieves a gauge value from the database.
func (s *DBStorage) GetGauge(name string) (float64, error) {
	var v float64
	err := retrier.Do(context.Background(), func() error {
		return s.pool.QueryRow(context.Background(), `SELECT value FROM gauges WHERE id=$1`, name).Scan(&v)
	}, isPGConnError, retrier.DefaultDelays)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrMetricNotFound
	}
	return v, err
}

// GetCounter retrieves a counter value from the database.
func (s *DBStorage) GetCounter(name string) (int64, error) {
	var v int64
	err := retrier.Do(context.Background(), func() error {
		return s.pool.QueryRow(context.Background(), `SELECT value FROM counters WHERE id=$1`, name).Scan(&v)
	}, isPGConnError, retrier.DefaultDelays)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrMetricNotFound
	}
	return v, err
}

// SetGauge overwrites a gauge value in the database.
func (s *DBStorage) SetGauge(name string, value float64) {
	_ = retrier.Do(context.Background(), func() error {
		_, err := s.pool.Exec(context.Background(), sqlSetGauges,
			name, value,
		)
		return err
	}, isPGConnError, retrier.DefaultDelays)
}

// SetCounter overwrites a counter value in the database.
func (s *DBStorage) SetCounter(name string, value int64) {
	_ = retrier.Do(context.Background(), func() error {
		_, err := s.pool.Exec(context.Background(), sqlSetCounters,
			name, value,
		)
		return err
	}, isPGConnError, retrier.DefaultDelays)
}

// AllGauges returns all gauge metrics stored in the database.
func (s *DBStorage) AllGauges() map[string]float64 {
	var rows pgx.Rows
	err := retrier.Do(context.Background(), func() error {
		var e error
		rows, e = s.pool.Query(context.Background(), `SELECT id, value FROM gauges`)
		return e
	}, isPGConnError, retrier.DefaultDelays)

	if err != nil {
		return map[string]float64{}
	}
	defer rows.Close()
	res := make(map[string]float64)
	for rows.Next() {
		var id string
		var val float64
		if err := rows.Scan(&id, &val); err == nil {
			res[id] = val
		}
	}
	return res
}

// AllCounters returns all counter metrics stored in the database.
func (s *DBStorage) AllCounters() map[string]int64 {
	var rows pgx.Rows
	err := retrier.Do(context.Background(), func() error {
		var e error
		rows, e = s.pool.Query(context.Background(), `SELECT id, value FROM counters`)
		return e
	}, isPGConnError, retrier.DefaultDelays)
	if err != nil {
		return map[string]int64{}
	}
	defer rows.Close()
	res := make(map[string]int64)
	for rows.Next() {
		var id string
		var val int64
		if err := rows.Scan(&id, &val); err == nil {
			res[id] = val
		}
	}
	return res
}

// UpdateBatch performs a batch upsert of metrics in a single transaction.
func (s *DBStorage) UpdateBatch(metrics []models.Metrics) (err error) {
	if len(metrics) == 0 {
		return nil
	}

	gm, cm := aggregateMetrics(metrics)
	if len(gm) == 0 && len(cm) == 0 {
		return nil
	}

	ctx := context.Background()
	var tx pgx.Tx
	if err = retrier.Do(ctx, func() error {
		var e error
		tx, e = s.pool.Begin(ctx)
		return e
	}, isPGConnError, retrier.DefaultDelays); err != nil {
		return err
	}
	defer s.commitOrRollback(ctx, tx, &err)

	if len(gm) > 0 {
		ids, vals := mapToSlices(gm)
		if err = execUpsertGauges(ctx, tx, ids, vals); err != nil {
			return err
		}
	}
	if len(cm) > 0 {
		ids, vals := mapToSlices(cm)
		if err = execUpsertCounters(ctx, tx, ids, vals); err != nil {
			return err
		}
	}
	return nil
}

func aggregateMetrics(metrics []models.Metrics) (map[string]float64, map[string]int64) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	for i := range metrics {
		m := metrics[i]
		switch m.MType {
		case models.GaugeType:
			if m.Value != nil {
				gauges[m.ID] = *m.Value
			}
		case models.CounterType:
			if m.Delta != nil {
				counters[m.ID] += *m.Delta
			}
		}
	}
	return gauges, counters
}

func execUpsertGauges(ctx context.Context, tx pgx.Tx, ids []string, values []float64) error {
	if len(ids) == 0 {
		return nil
	}
	return retrier.Do(ctx, func() error {
		_, err := tx.Exec(ctx, sqlUpsertGauges, ids, values)
		return err
	}, isPGConnError, retrier.DefaultDelays)
}

func execUpsertCounters(ctx context.Context, tx pgx.Tx, ids []string, values []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return retrier.Do(ctx, func() error {
		_, err := tx.Exec(ctx, sqlUpsertCounters, ids, values)
		return err
	}, isPGConnError, retrier.DefaultDelays)
}

func (s *DBStorage) commitOrRollback(ctx context.Context, tx pgx.Tx, errp *error) {
	if *errp != nil {
		_ = tx.Rollback(ctx)
		return
	}
	*errp = retrier.Do(ctx, func() error {
		return tx.Commit(ctx)
	}, isPGConnError, retrier.DefaultDelays)
}

func mapToSlices[V any](m map[string]V) ([]string, []V) {
	ids := make([]string, 0, len(m))
	vals := make([]V, 0, len(m))
	for id, v := range m {
		ids = append(ids, id)
		vals = append(vals, v)
	}
	return ids, vals
}

func isPGConnError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown,
			pgerrcode.ProtocolViolation:
			return true
		}
	}
	return false
}

var _ MetricStorage = NewDBStorage(nil)
