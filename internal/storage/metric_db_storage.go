package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
)

type DBStorage struct {
	pool db.Pool
}

func NewDBStorage(p db.Pool) *DBStorage {
	return &DBStorage{pool: p}
}

func (s *DBStorage) UpdateGauge(name string, value float64) {
	_, _ = s.pool.Exec(context.Background(),
		`INSERT INTO gauges(id, value, updated_at) VALUES ($1,$2,NOW())
        ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
		name, value,
	)
}

func (s *DBStorage) UpdateCounter(name string, delta int64) {
	_, _ = s.pool.Exec(context.Background(),
		`INSERT INTO counters(id, value, updated_at) VALUES ($1,$2,NOW())
        ON CONFLICT (id) DO UPDATE SET value = counters.value + EXCLUDED.value, updated_at = NOW()`,
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
	_, _ = s.pool.Exec(context.Background(),
		`INSERT INTO gauges(id, value, updated_at) VALUES ($1,$2,NOW())
        ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
		name, value,
	)
}

func (s *DBStorage) SetCounter(name string, value int64) {
	_, _ = s.pool.Exec(context.Background(),
		`INSERT INTO counters(id, value, updated_at) VALUES ($1,$2,NOW())
        ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
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

var _ MetricStorage = NewDBStorage(nil)
