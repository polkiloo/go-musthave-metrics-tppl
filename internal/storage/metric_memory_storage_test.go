package storage

import (
	"sync"
	"testing"
)

func TestNewMemStorage_InitialState(t *testing.T) {
	ms := NewMemStorage()
	ms.mu.RLock()
	if ms.gauges == nil {
		t.Error("gauges map should be initialized, but got nil")
	}
	if ms.counters == nil {
		t.Error("counters map should be initialized, but got nil")
	}
	ms.mu.RUnlock()
}

func TestUpdateGaugeAndGetGauge(t *testing.T) {
	ms := NewMemStorage()
	if _, ok := ms.GetGauge("g1"); ok {
		t.Error("expected gauge 'g1' to not exist initially")
	}
	ms.UpdateGauge("g1", 42.5)
	v, ok := ms.GetGauge("g1")
	if !ok {
		t.Error("expected gauge 'g1' to exist after update")
	}
	if v != 42.5 {
		t.Errorf("got gauge value %v; want 42.5", v)
	}
}

func TestUpdateCounterAndGetCounter(t *testing.T) {
	ms := NewMemStorage()
	if _, ok := ms.GetCounter("c1"); ok {
		t.Error("expected counter 'c1' to not exist initially")
	}
	ms.UpdateCounter("c1", 1)
	ms.UpdateCounter("c1", 4)
	v, ok := ms.GetCounter("c1")
	if !ok {
		t.Error("expected counter 'c1' to exist after updates")
	}
	if v != 5 {
		t.Errorf("got counter value %d; want 5", v)
	}
}

func TestConcurrentGaugeUpdates(t *testing.T) {
	ms := NewMemStorage()
	var wg sync.WaitGroup
	n := 1000
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			ms.UpdateGauge("gauge", float64(i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			ms.UpdateGauge("gauge", float64(n-i))
		}
	}()
	wg.Wait()
	v, ok := ms.GetGauge("gauge")
	if !ok {
		t.Fatal("expected gauge 'gauge' to exist after concurrent updates")
	}
	if v < 0 || v > float64(n) {
		t.Errorf("gauge value out of expected range: got %v", v)
	}
}

func TestConcurrentCounterUpdates(t *testing.T) {
	ms := NewMemStorage()
	var wg sync.WaitGroup
	n := 1000
	nGo := 5
	wg.Add(nGo)
	for g := 0; g < nGo; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				ms.UpdateCounter("counter", 1)
			}
		}()
	}
	wg.Wait()
	expected := int64(nGo * n)
	v, ok := ms.GetCounter("counter")
	if !ok {
		t.Fatal("expected counter 'counter' to exist after concurrent updates")
	}
	if v != expected {
		t.Errorf("got counter value %d; want %d", v, expected)
	}
}
