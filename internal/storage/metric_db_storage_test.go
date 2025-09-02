package storage

import (
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

func TestUpdateGauge_PositiveAndError_NoPanic(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	s := NewDBStorage(mock)

	mock.ExpectExec(regexp.QuoteMeta(sqlUpdateGauges)).
		WithArgs("Temp", 12.34).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	s.UpdateGauge("Temp", 12.34)

	mock.ExpectExec(regexp.QuoteMeta(sqlUpdateGauges)).
		WithArgs("Temp", 1.0).
		WillReturnError(errors.New("boom"))
	s.UpdateGauge("Temp", 1.0)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateCounter_PositiveAndError_NoPanic(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	s := NewDBStorage(mock)

	mock.ExpectExec(regexp.QuoteMeta(sqlUpdateCounters)).
		WithArgs("Poll", int64(5)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	s.UpdateCounter("Poll", 5)

	mock.ExpectExec(regexp.QuoteMeta(sqlUpdateCounters)).
		WithArgs("Poll", int64(1)).
		WillReturnError(errors.New("fail"))
	s.UpdateCounter("Poll", 1)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSetGauge_PositiveAndError_NoPanic(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	s := NewDBStorage(mock)

	mock.ExpectExec(regexp.QuoteMeta(sqlSetGauges)).
		WithArgs("G", 1.1).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	s.SetGauge("G", 1.1)

	mock.ExpectExec(regexp.QuoteMeta(sqlSetGauges)).
		WithArgs("G", 2.2).
		WillReturnError(errors.New("fail"))
	s.SetGauge("G", 2.2)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestSetCounter_PositiveAndError_NoPanic(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	s := NewDBStorage(mock)

	mock.ExpectExec(regexp.QuoteMeta(sqlSetCounters)).
		WithArgs("C", int64(7)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	s.SetCounter("C", 7)

	mock.ExpectExec(regexp.QuoteMeta(sqlSetCounters)).
		WithArgs("C", int64(8)).
		WillReturnError(errors.New("fail"))
	s.SetCounter("C", 8)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestGetGauge_Positive_NotFound_DBError(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	rows := pgxmock.NewRows([]string{"value"}).AddRow(3.14)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT value FROM gauges WHERE id=$1`)).
		WithArgs("Temp").WillReturnRows(rows)
	v, err := s.GetGauge("Temp")
	if err != nil || v != 3.14 {
		t.Fatalf("want 3.14,nil got %v,%v", v, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT value FROM gauges WHERE id=$1`)).
		WithArgs("X").WillReturnError(pgx.ErrNoRows)
	_, err = s.GetGauge("X")
	if !errors.Is(err, ErrMetricNotFound) {
		t.Fatalf("want ErrMetricNotFound, got %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT value FROM gauges WHERE id=$1`)).
		WithArgs("Y").WillReturnError(errors.New("db down"))
	_, err = s.GetGauge("Y")
	if err == nil || errors.Is(err, ErrMetricNotFound) {
		t.Fatalf("want raw db error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetCounter_Positive_NotFound_DBError(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	rows := pgxmock.NewRows([]string{"value"}).AddRow(int64(42))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT value FROM counters WHERE id=$1`)).
		WithArgs("Poll").WillReturnRows(rows)
	v, err := s.GetCounter("Poll")
	if err != nil || v != 42 {
		t.Fatalf("want 42,nil got %v,%v", v, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT value FROM counters WHERE id=$1`)).
		WithArgs("X").WillReturnError(pgx.ErrNoRows)
	_, err = s.GetCounter("X")
	if !errors.Is(err, ErrMetricNotFound) {
		t.Fatalf("want ErrMetricNotFound")
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT value FROM counters WHERE id=$1`)).
		WithArgs("Y").WillReturnError(errors.New("db err"))
	_, err = s.GetCounter("Y")
	if err == nil || errors.Is(err, ErrMetricNotFound) {
		t.Fatalf("want raw db error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAllGauges_Positive_QueryErr_ScanErr(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	rows := pgxmock.NewRows([]string{"id", "value"}).
		AddRow("a", 1.1).AddRow("b", 2.2)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, value FROM gauges`)).
		WillReturnRows(rows)
	got := s.AllGauges()
	if len(got) != 2 || got["a"] != 1.1 || got["b"] != 2.2 {
		t.Fatalf("unexpected map: %+v", got)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, value FROM gauges`)).
		WillReturnError(errors.New("query fail"))
	got = s.AllGauges()
	if len(got) != 0 {
		t.Fatalf("expected empty on error")
	}

	rowsBad := pgxmock.NewRows([]string{"id", "value"}).
		AddRow("ok", 3.3).
		AddRow("bad", "oops")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, value FROM gauges`)).
		WillReturnRows(rowsBad)
	got = s.AllGauges()
	if len(got) != 1 || got["ok"] != 3.3 {
		t.Fatalf("unexpected after scan error: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAllCounters_Positive_QueryErr_ScanErr(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	rows := pgxmock.NewRows([]string{"id", "value"}).
		AddRow("x", int64(10)).
		AddRow("y", int64(20))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, value FROM counters`)).
		WillReturnRows(rows)
	got := s.AllCounters()
	if len(got) != 2 || got["x"] != 10 || got["y"] != 20 {
		t.Fatalf("unexpected map: %+v", got)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, value FROM counters`)).
		WillReturnError(errors.New("query fail"))
	got = s.AllCounters()
	if len(got) != 0 {
		t.Fatalf("expected empty on error")
	}

	rowsBad := pgxmock.NewRows([]string{"id", "value"}).
		AddRow("ok", int64(1)).
		AddRow("bad", "oops")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, value FROM counters`)).
		WillReturnRows(rowsBad)
	got = s.AllCounters()
	if len(got) != 1 || got["ok"] != 1 {
		t.Fatalf("unexpected after scan error: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateBatch_Positive_BothGroups(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(sqlUpsertGauges)).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))
	mock.ExpectExec(regexp.QuoteMeta(sqlUpsertCounters)).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))
	mock.ExpectCommit()

	err := s.UpdateBatch([]models.Metrics{
		{ID: "g1", MType: models.GaugeType, Value: pFloat64(1)},
		{ID: "c1", MType: models.CounterType, Delta: pInt64(2)},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateBatch_BeginError(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	mock.ExpectBegin().WillReturnError(errors.New("begin fail"))
	err := s.UpdateBatch([]models.Metrics{{ID: "g", MType: models.GaugeType, Value: pFloat64(1)}})
	if err == nil {
		t.Fatalf("want begin error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateBatch_GaugesExecError_Rollback(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(sqlUpsertGauges)).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(errors.New("upsert gauges fail"))
	mock.ExpectRollback()

	err := s.UpdateBatch([]models.Metrics{{ID: "g", MType: models.GaugeType, Value: pFloat64(1)}})
	if err == nil {
		t.Fatalf("want gauges exec error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateBatch_CountersExecError_Rollback(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(sqlUpsertCounters)).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(errors.New("counters fail"))
	mock.ExpectRollback()

	err := s.UpdateBatch([]models.Metrics{{ID: "c", MType: models.CounterType, Delta: pInt64(1)}})
	if err == nil {
		t.Fatalf("want counters exec error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateBatch_CommitError(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	s := NewDBStorage(mock)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(sqlUpsertGauges)).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err := s.UpdateBatch([]models.Metrics{{ID: "g1", MType: models.GaugeType, Value: pFloat64(1)}})
	if err == nil || err.Error() != "commit failed" {
		t.Fatalf("want commit failed, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAggregateMetrics(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "g1", MType: models.GaugeType, Value: pFloat64(1)},
		{ID: "g1", MType: models.GaugeType, Value: pFloat64(2)},
		{ID: "c1", MType: models.CounterType, Delta: pInt64(5)},
		{ID: "c1", MType: models.CounterType, Delta: pInt64(3)},
		{ID: "cNil", MType: models.CounterType, Delta: nil},
		{ID: "gNil", MType: models.GaugeType, Value: nil},
	}
	g, c := aggregateMetrics(metrics)
	if g["g1"] != 2 || len(g) != 1 {
		t.Fatalf("gauges wrong: %+v", g)
	}
	if c["c1"] != 8 || len(c) != 1 {
		t.Fatalf("counters wrong: %+v", c)
	}
}

func TestMapToSlices(t *testing.T) {
	mf := map[string]float64{"a": 1, "b": 2}
	ids, vals := mapToSlices(mf)
	if len(ids) != 2 || len(vals) != 2 {
		t.Fatalf("bad len")
	}
	pairs := map[string]float64{}
	for i := range ids {
		pairs[ids[i]] = vals[i]
	}
	if !reflect.DeepEqual(pairs, mf) {
		t.Fatalf("pairs mismatch")
	}

	mi := map[string]int64{}
	ids2, vals2 := mapToSlices(mi)
	if len(ids2) != 0 || len(vals2) != 0 {
		t.Fatalf("expected empties")
	}
}

func pInt64(v int64) *int64       { return &v }
func pFloat64(v float64) *float64 { return &v }
