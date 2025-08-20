package service

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

func TestProcessUpdate_Gauge(t *testing.T) {
	svc := NewMetricService()
	m := &models.Metrics{ID: "g1", MType: models.GaugeType, Value: Float64Ptr(3.14)}
	err := svc.ProcessUpdate(m)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessUpdate_Counter(t *testing.T) {
	svc := NewMetricService()

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
	svc := NewMetricService()
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
	svc := NewMetricService()
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
	svc := NewMetricService()
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
	svc := NewMetricService()
	_, err := svc.ProcessGetValue("not_exist", models.GaugeType)
	if err == nil {
		t.Errorf("expected error for missing gauge, got nil")
	}
}

func TestMetricService_SaveLoadFile(t *testing.T) {
	svc := NewMetricService()
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

	svc2 := NewMetricService()
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
	svc := NewMetricService()
	if err := svc.SaveFile(""); err != nil {
		t.Fatalf("SaveFile empty path: %v", err)
	}
	if err := svc.LoadFile(""); err != nil {
		t.Fatalf("LoadFile empty path: %v", err)
	}
}

func Float64Ptr(v float64) *float64 { return &v }
func Int64Ptr(v int64) *int64       { return &v }
