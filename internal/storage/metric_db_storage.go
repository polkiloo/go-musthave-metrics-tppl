package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
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

type DBStorage struct {
	pool db.Pool
}

func NewDBStorage(p db.Pool) *DBStorage {
	return &DBStorage{pool: p}
}

func (s *DBStorage) UpdateGauge(name string, value float64) {
	_, _ = s.pool.Exec(context.Background(), sqlUpdateGauges,
		name, value,
	)
}

func (s *DBStorage) UpdateCounter(name string, delta int64) {
	_, _ = s.pool.Exec(context.Background(), sqlUpdateCounters,
		name, delta,
	)
}

func (s *DBStorage) GetGauge(name string) (float64, error) {
	var v float64
	err := s.pool.QueryRow(context.Background(), `SELECT value FROM gauges WHERE id=$1`, name).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrMetricNotFound
	}
	return v, err
}

func (s *DBStorage) GetCounter(name string) (int64, error) {
	var v int64
	err := s.pool.QueryRow(context.Background(), `SELECT value FROM counters WHERE id=$1`, name).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrMetricNotFound
	}
	return v, err
}

func (s *DBStorage) SetGauge(name string, value float64) {
	_, _ = s.pool.Exec(context.Background(), sqlSetGauges,
		name, value,
	)
}

func (s *DBStorage) SetCounter(name string, value int64) {
	_, _ = s.pool.Exec(context.Background(), sqlSetCounters,
		name, value,
	)
}

func (s *DBStorage) AllGauges() map[string]float64 {
	rows, err := s.pool.Query(context.Background(), `SELECT id, value FROM gauges`)
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

func (s *DBStorage) AllCounters() map[string]int64 {
	rows, err := s.pool.Query(context.Background(), `SELECT id, value FROM counters`)
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

func (s *DBStorage) UpdateBatch(metrics []models.Metrics) (err error) {
	if len(metrics) == 0 {
		return nil
	}

	gm, cm := aggregateMetrics(metrics)
	if len(gm) == 0 && len(cm) == 0 {
		return nil
	}

	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
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
	_, err := tx.Exec(ctx, sqlUpsertGauges, ids, values)
	return err
}

func execUpsertCounters(ctx context.Context, tx pgx.Tx, ids []string, values []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx, sqlUpsertCounters, ids, values)
	return err
}

func (s *DBStorage) commitOrRollback(ctx context.Context, tx pgx.Tx, errp *error) {
	if *errp != nil {
		_ = tx.Rollback(ctx)
		return
	}
	*errp = tx.Commit(ctx)
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

var _ MetricStorage = NewDBStorage(nil)
