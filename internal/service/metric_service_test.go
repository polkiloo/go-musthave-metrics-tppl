package service

import (
	"strconv"
	"sync"
	"testing"
)

func TestProcessUpdate_Gauge(t *testing.T) {
	svc := NewMetricService()

	err := svc.ProcessUpdate("gauge", "g1", "3.14")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = svc.ProcessUpdate("gauge", "g2", "notfloat")
	if err == nil {
		t.Errorf("expected error on invalid gauge value, got nil")
	}
}

func TestProcessUpdate_Counter(t *testing.T) {
	svc := NewMetricService()

	err := svc.ProcessUpdate("counter", "c1", "10")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = svc.ProcessUpdate("counter", "c1", "5")
	if err != nil {
		t.Fatalf("expected no error on accumulation, got %v", err)
	}

	err = svc.ProcessUpdate("counter", "c2", "notint")
	if err == nil {
		t.Errorf("expected error on invalid counter value, got nil")
	}
}

func TestProcessUpdate_UnknownType(t *testing.T) {
	svc := NewMetricService()
	err := svc.ProcessUpdate("unknown", "x", "1")
	if err == nil {
		t.Errorf("expected error on unknown metric type, got nil")
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
			err := svc.ProcessUpdate("gauge", "gauge", strconv.Itoa(i)+".0")
			if err != nil {
				t.Errorf("gauge update error: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			err := svc.ProcessUpdate("counter", "counter", "1")
			if err != nil {
				t.Errorf("counter update error: %v", err)
			}
		}
	}()

	wg.Wait()
}
