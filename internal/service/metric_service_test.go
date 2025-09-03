package service

import (
	"errors"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func TestProcessUpdate_Gauge(t *testing.T) {
	svc := NewMetricService(storage.NewMemStorage())
	m := &models.Metrics{ID: "g1", MType: models.GaugeType, Value: Float64Ptr(3.14)}
	err := svc.ProcessUpdate(m)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessUpdate_Counter(t *testing.T) {
	svc := NewMetricService(storage.NewMemStorage())

	m := &models.Metrics{ID: "c1", MType: models.CounterType, Delta: Int64Ptr(10)}
	err := svc.ProcessUpdate(m)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	m = &models.Metrics{ID: "c1", MType: models.CounterType, Delta: Int64Ptr(5)}
	err = svc.ProcessUpdate(m)
	if err != nil {
		t.Fatalf("expected no error on accumulation, got %v", err)
	}
}

func TestProcessUpdate_Concurrent(t *testing.T) {
	svc := NewMetricService(storage.NewMemStorage())
	var wg sync.WaitGroup

	n := 1000
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			f := float64(i)
			m, _ := models.NewGaugeMetrics(models.GaugeNames[0], &f)
			err := svc.ProcessUpdate(m)
			if err != nil {
				t.Errorf("gauge update error: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			j := int64(i)
			m, _ := models.NewCounterMetrics(models.CounterNames[0], &j)
			err := svc.ProcessUpdate(m)
			if err != nil {
				t.Errorf("counter update error: %v", err)
			}
		}
	}()

	wg.Wait()
}

func TestProcessGet_Gauge(t *testing.T) {
	svc := NewMetricService(storage.NewMemStorage())
	f := float64(2.71)
	m, _ := models.NewGaugeMetrics(models.GaugeNames[0], &f)
	_ = svc.ProcessUpdate(m)

	val, err := svc.ProcessGetValue(m.ID, m.MType)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if *val.Value != f {
		t.Errorf("expected value '%f', got %f", f, *val.Value)
	}
}

func TestProcessGet_Counter(t *testing.T) {
	svc := NewMetricService(storage.NewMemStorage())
	j := int64(42)
	m, _ := models.NewCounterMetrics(models.CounterNames[0], &j)
	_ = svc.ProcessUpdate(m)

	val, err := svc.ProcessGetValue(m.ID, m.MType)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if *val.Delta != j {
		t.Errorf("expected value '%d', got %d", j, *val.Delta)
	}
}

func TestProcessGet_NotFound(t *testing.T) {
	svc := NewMetricService(storage.NewMemStorage())
	_, err := svc.ProcessGetValue("not_exist", models.GaugeType)
	if err == nil {
		t.Errorf("expected error for missing gauge, got nil")
	}
}

func TestMetricService_SaveLoadFile(t *testing.T) {
	svc := NewMetricService(storage.NewMemStorage())
	f := float64(1.23)
	g, _ := models.NewGaugeMetrics("g", &f)
	_ = svc.ProcessUpdate(g)
	cval := int64(7)
	c, _ := models.NewCounterMetrics("c", &cval)
	_ = svc.ProcessUpdate(c)

	tmp := filepath.Join(t.TempDir(), "m.json")
	if err := svc.SaveFile(tmp); err != nil {
		t.Fatalf("SaveFile error: %v", err)
	}

	svc2 := NewMetricService(storage.NewMemStorage())
	if err := svc2.LoadFile(tmp); err != nil {
		t.Fatalf("LoadFile error: %v", err)
	}
	gv, err := svc2.ProcessGetValue("g", models.GaugeType)
	if err != nil || *gv.Value != f {
		t.Fatalf("gauge mismatch: %v %v", gv, err)
	}
	cv, err := svc2.ProcessGetValue("c", models.CounterType)
	if err != nil || *cv.Delta != cval {
		t.Fatalf("counter mismatch: %v %v", cv, err)
	}
}

func TestMetricService_SaveLoadFile_EmptyPath(t *testing.T) {
	svc := NewMetricService(storage.NewMemStorage())
	if err := svc.SaveFile(""); err != nil {
		t.Fatalf("SaveFile empty path: %v", err)
	}
	if err := svc.LoadFile(""); err != nil {
		t.Fatalf("LoadFile empty path: %v", err)
	}
}

func TestProcessUpdates_EmptySlice(t *testing.T) {
	s := &MetricService{store: test.NewFakeStorage()}
	if err := s.ProcessUpdates(nil); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
	if err := s.ProcessUpdates([]models.Metrics{}); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestProcessUpdates_Batch_Success(t *testing.T) {
	base := test.NewFakeStorage()
	fb := &test.FakeBatchStore{FakeStorage: base}
	s := &MetricService{store: fb}

	in := []models.Metrics{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	if err := s.ProcessUpdates(in); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
	if !reflect.DeepEqual(fb.Got, in) {
		t.Fatalf("batch got != in: %#v != %#v", fb.Got, in)
	}
}

func TestProcessUpdates_Batch_Error(t *testing.T) {
	base := test.NewFakeStorage()
	wantErr := errors.New("boom")
	fb := &test.FakeBatchStore{FakeStorage: base, Err: wantErr}
	s := &MetricService{store: fb}

	in := []models.Metrics{{ID: "x"}}
	err := s.ProcessUpdates(in)
	if !errors.Is(err, wantErr) {
		t.Fatalf("want %v, got %v", wantErr, err)
	}
}

func TestProcessUpdates_Fallback_Success(t *testing.T) {
	old := processUpdateFn
	t.Cleanup(func() { processUpdateFn = old })

	var called []string
	processUpdateFn = func(_ *MetricService, m *models.Metrics) error {
		called = append(called, m.ID)
		return nil
	}

	s := &MetricService{store: &test.FakeNoBatchStore{FakeStorage: test.NewFakeStorage()}}
	in := []models.Metrics{{ID: "1"}, {ID: "2"}, {ID: "3"}}

	if err := s.ProcessUpdates(in); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
	want := []string{"1", "2", "3"}
	if !reflect.DeepEqual(called, want) {
		t.Fatalf("update order mismatch: got %v, want %v", called, want)
	}
}

func TestProcessUpdates_Fallback_StopsOnFirstError(t *testing.T) {
	old := processUpdateFn
	t.Cleanup(func() { processUpdateFn = old })

	failAt := 1
	var calls int
	wantErr := errors.New("fail")

	processUpdateFn = func(_ *MetricService, _ *models.Metrics) error {
		if calls == failAt {
			calls++
			return wantErr
		}
		calls++
		return nil
	}

	s := &MetricService{store: &test.FakeNoBatchStore{FakeStorage: test.NewFakeStorage()}}
	in := []models.Metrics{{ID: "a"}, {ID: "b"}, {ID: "c"}}

	err := s.ProcessUpdates(in)
	if !errors.Is(err, wantErr) {
		t.Fatalf("want %v, got %v", wantErr, err)
	}
	if calls != failAt+1 {
		t.Fatalf("should stop after first error: calls=%d, want=%d", calls, failAt+1)
	}
}

func Float64Ptr(v float64) *float64 { return &v }
func Int64Ptr(v int64) *int64       { return &v }
