package storage

import (
	"sync"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

func TestMemStorage_UpdateAndGetGauge(t *testing.T) {
	m := NewMemStorage()
	m.UpdateGauge("cpu", 2.3)
	val, err := m.GetGauge("cpu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 2.3 {
		t.Errorf("got %v, want %v", val, 2.3)
	}
}

func TestMemStorage_GetGauge_NotFound(t *testing.T) {
	m := NewMemStorage()
	_, err := m.GetGauge("notfound")
	if err != ErrMetricNotFound {
		t.Errorf("expected ErrMetricNotFound, got %v", err)
	}
}

func TestMemStorage_UpdateAndGetCounter(t *testing.T) {
	m := NewMemStorage()
	m.UpdateCounter("hits", 5)
	val, err := m.GetCounter("hits")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 5 {
		t.Errorf("got %v, want %v", val, 5)
	}
}

func TestMemStorage_UpdateCounter_Accumulation(t *testing.T) {
	m := NewMemStorage()
	m.UpdateCounter("hits", 2)
	m.UpdateCounter("hits", 3)
	val, err := m.GetCounter("hits")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 5 {
		t.Errorf("got %v, want %v", val, 5)
	}
}

func TestMemStorage_GetCounter_NotFound(t *testing.T) {
	m := NewMemStorage()
	_, err := m.GetCounter("nope")
	if err != ErrMetricNotFound {
		t.Errorf("expected ErrMetricNotFound, got %v", err)
	}
}

func TestMemStorage_OverwriteGauge(t *testing.T) {
	m := NewMemStorage()
	m.UpdateGauge("temp", 1.0)
	m.UpdateGauge("temp", 7.5)
	val, err := m.GetGauge("temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 7.5 {
		t.Errorf("got %v, want %v", val, 7.5)
	}
}

func TestMemStorage_ConcurrentAccess(t *testing.T) {
	m := NewMemStorage()
	const n = 100
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val float64) {
			defer wg.Done()
			m.UpdateGauge("g", val)
		}(float64(i))
	}
	wg.Wait()
	v, err := m.GetGauge("g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v < 0 || v >= float64(n) {
		t.Errorf("unexpected gauge value after concurrency: %v", v)
	}

	wg = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.UpdateCounter("c", 1)
		}()
	}
	wg.Wait()
	count, err := m.GetCounter("c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != int64(n) {
		t.Errorf("counter got %v, want %v", count, n)
	}
}

func TestNumMemStorage_AddFloatAndInt(t *testing.T) {
	g := NewNumMemStorage[float64]()
	g.Add("f", 1.2)
	g.Add("f", 2.3)
	val, err := g.Get("f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 3.5 {
		t.Errorf("got %v, want 3.5", val)
	}

	c := NewNumMemStorage[int64]()
	c.Add("i", 2)
	c.Add("i", 5)
	ival, err := c.Get("i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ival != 7 {
		t.Errorf("got %v, want 7", ival)
	}
}

func TestMemStorage_InitialState(t *testing.T) {
	m := NewMemStorage()
	_, err := m.GetGauge("new")
	if err != ErrMetricNotFound {
		t.Errorf("expected ErrMetricNotFound for new gauge, got %v", err)
	}
	_, err = m.GetCounter("new")
	if err != ErrMetricNotFound {
		t.Errorf("expected ErrMetricNotFound for new counter, got %v", err)
	}
}

func TestMemStorage_AllGaugesAndCounters(t *testing.T) {
	m := NewMemStorage()
	m.UpdateGauge("g1", 1.5)
	m.UpdateCounter("c1", 3)
	gs := m.AllGauges()
	cs := m.AllCounters()
	if gs["g1"] != 1.5 || cs["c1"] != 3 {
		t.Fatalf("unexpected maps: %v %v", gs, cs)
	}
	gs["g1"] = 99
	cs["c1"] = 99
	g, _ := m.GetGauge("g1")
	c, _ := m.GetCounter("c1")
	if g != 1.5 || c != 3 {
		t.Fatalf("maps not copies: g=%v c=%v", g, c)
	}
}

func TestMemStorage_Snapshot(t *testing.T) {
	m := NewMemStorage()
	m.UpdateGauge("g", 2.5)
	m.UpdateCounter("c", 7)

	snapshot := m.Snapshot()
	if len(snapshot) != 2 {
		t.Fatalf("snapshot length = %d, want 2", len(snapshot))
	}

	values := make(map[string]float64)
	counters := make(map[string]int64)
	for _, metric := range snapshot {
		switch metric.MType {
		case models.GaugeType:
			if metric.Value == nil {
				t.Fatalf("gauge %q has nil value", metric.ID)
			}
			values[metric.ID] = *metric.Value
			*metric.Value = 99.0
		case models.CounterType:
			if metric.Delta == nil {
				t.Fatalf("counter %q has nil delta", metric.ID)
			}
			counters[metric.ID] = *metric.Delta
			*metric.Delta = 42
		}
	}

	if values["g"] != 2.5 {
		t.Fatalf("snapshot gauge value = %v, want 2.5", values["g"])
	}
	if counters["c"] != 7 {
		t.Fatalf("snapshot counter value = %v, want 7", counters["c"])
	}

	gv, _ := m.GetGauge("g")
	cv, _ := m.GetCounter("c")
	if gv != 2.5 || cv != 7 {
		t.Fatalf("mutating snapshot should not affect storage, got gauge=%v counter=%v", gv, cv)
	}
}
func TestMemStorage_SetCounter(t *testing.T) {
	m := NewMemStorage()
	m.UpdateCounter("c", 5)
	m.SetCounter("c", 10)
	v, err := m.GetCounter("c")
	if err != nil || v != 10 {
		t.Fatalf("SetCounter failed: %v %v", v, err)
	}
}

func TestMemStorage_SetGauge(t *testing.T) {
	m := NewMemStorage()
	m.UpdateGauge("c", 5.443)
	m.SetGauge("c", 1.23)
	v, err := m.GetGauge("c")
	if err != nil || v != 1.23 {
		t.Fatalf("SetGauge failed: %v %v", v, err)
	}
}

func TestMemStorage_UpdateBatch_MixedAndUnknownTypes(t *testing.T) {
	s := NewMemStorage()

	s.SetGauge("g_keep", 10.0)
	s.SetCounter("c_keep", 5)
	s.SetCounter("c_inc", 10)

	gv := 1.5
	d2 := int64(2)
	d3 := int64(3)

	in := []models.Metrics{
		{ID: "g1", MType: models.GaugeType, Value: &gv},
		{ID: "g_nil", MType: models.GaugeType, Value: nil},
		{ID: "c1", MType: models.CounterType, Delta: &d2},
		{ID: "c_nil", MType: models.CounterType, Delta: nil},
		{ID: "c_inc", MType: models.CounterType, Delta: &d3},
		{ID: "zzz", MType: models.MetricType("unknown")},
	}

	if err := s.UpdateBatch(in); err != nil {
		t.Fatalf("UpdateBatch() want nil error, got %v", err)
	}

	if got, _ := s.GetGauge("g1"); got != gv {
		t.Fatalf("GetGauge(g1) = %v, want %v", got, gv)
	}
	if got, _ := s.GetGauge("g_keep"); got != 10.0 {
		t.Fatalf("GetGauge(g_keep) = %v, want %v", got, 10.0)
	}
	if got, _ := s.GetCounter("c1"); got != 2 {
		t.Fatalf("GetCounter(c1) = %d, want %d", got, 2)
	}
	if got, _ := s.GetCounter("c_inc"); got != 13 {
		t.Fatalf("GetCounter(c_inc) = %d, want %d", got, 13)
	}
	if got, _ := s.GetCounter("c_keep"); got != 5 {
		t.Fatalf("GetCounter(c_keep) = %d, want %d", got, 5)
	}
}

func TestMemStorage_UpdateBatch_EmptyInput(t *testing.T) {
	var s MemStorage
	if err := s.UpdateBatch(nil); err != nil {
		t.Fatalf("UpdateBatch(nil) want nil, got %v", err)
	}
	if err := s.UpdateBatch([]models.Metrics{}); err != nil {
		t.Fatalf("UpdateBatch(empty) want nil, got %v", err)
	}
}
