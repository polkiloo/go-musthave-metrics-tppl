package storage

import (
	"sync"
	"testing"
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
