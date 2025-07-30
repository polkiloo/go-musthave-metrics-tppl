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

func TestProcessGet_Gauge(t *testing.T) {
	svc := NewMetricService()
	_ = svc.ProcessUpdate("gauge", "g1", "2.71")

	val, err := svc.ProcessGetValue("gauge", "g1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "2.71" {
		t.Errorf("expected value '2.71', got %q", val)
	}
}

func TestProcessGet_Counter(t *testing.T) {
	svc := NewMetricService()
	_ = svc.ProcessUpdate("counter", "c1", "42")

	val, err := svc.ProcessGetValue("counter", "c1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "42" {
		t.Errorf("expected value '42', got %q", val)
	}
}

func TestProcessGet_UnknownType(t *testing.T) {
	svc := NewMetricService()
	_, err := svc.ProcessGetValue("unknown", "x")
	if err == nil {
		t.Errorf("expected error for unknown metric type, got nil")
	}
}

func TestProcessGet_NotFound(t *testing.T) {
	svc := NewMetricService()
	_, err := svc.ProcessGetValue("gauge", "not_exist")
	if err == nil {
		t.Errorf("expected error for missing gauge, got nil")
	}
	_, err = svc.ProcessGetValue("counter", "not_exist")
	if err == nil {
		t.Errorf("expected error for missing counter, got nil")
	}
}
