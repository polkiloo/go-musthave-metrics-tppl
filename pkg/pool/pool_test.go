package pool

import (
	"runtime"
	"testing"
)

type sample struct {
	Value        int
	ResetCounter *int
}

func newSample() *sample {
	return &sample{Value: 123, ResetCounter: new(int)}
}

func (s *sample) Reset() {
	s.Value = 0
	if s.ResetCounter != nil {
		*s.ResetCounter++
	}
}

func TestPoolGetAndPut(t *testing.T) {
	p := New(newSample)

	obj := p.Get()
	if obj.Value != 123 {
		t.Fatalf("expected constructor to initialize value, got %d", obj.Value)
	}

	obj.Value = 42
	p.Put(obj)

	if obj.Value != 0 {
		t.Fatalf("expected object to be reset before being put back, got %d", obj.Value)
	}
	if obj.ResetCounter == nil || *obj.ResetCounter != 1 {
		t.Fatalf("expected reset to be called once, got %v", obj.ResetCounter)
	}

	reused := p.Get()
	if reused.ResetCounter == nil || *reused.ResetCounter == 0 {
		t.Fatalf("expected pooled object to retain reset counter, got %v", reused.ResetCounter)
	}
}

// generate:reset
type valueReceiver struct{ called *int }

func newValueReceiver() valueReceiver {
	counter := 0
	return valueReceiver{called: &counter}
}

func (v valueReceiver) Reset() {
	if v.called != nil {
		*v.called++
	}
}

func TestPoolSupportsValueTypes(t *testing.T) {
	p := New(newValueReceiver)
	obj := p.Get()
	if obj.called == nil {
		t.Fatal("expected called counter to be initialized")
	}
	p.Put(obj)
	if *obj.called != 1 {
		t.Fatalf("expected reset to be invoked for value type, got %d", *obj.called)
	}
}

// generate:reset
type heavy struct {
	data []byte
}

func newHeavy() *heavy {
	return &heavy{data: make([]byte, 1<<16)}
}

func (h *heavy) Reset() {
	for i := range h.data {
		h.data[i] = 0
	}
}

// generate:reset
type panicSample struct{}

var _ Resettable = (*panicSample)(nil)

func (panicSample) Reset() { panic("reset failed") }

func TestPoolReusesObjectAfterPut(t *testing.T) {
	p := New(newSample)

	first := p.Get()
	first.Value = 55
	p.Put(first)

	second := p.Get()
	if first != second {
		t.Fatalf("expected pooled object to be reused")
	}
	if second.Value != 0 {
		t.Fatalf("expected value to be reset on reuse, got %d", second.Value)
	}
}

func TestPoolUsesConstructorWhenEmpty(t *testing.T) {
	p := New(newSample)

	first := p.Get()
	second := p.Get()

	if first == second {
		t.Fatalf("expected distinct instances when pool is empty")
	}
	if second.Value != 123 {
		t.Fatalf("expected constructor to set initial value, got %d", second.Value)
	}
}

func TestPoolTracksMultipleResets(t *testing.T) {
	p := New(newSample)
	obj := p.Get()

	p.Put(obj)
	p.Put(obj)

	if obj.ResetCounter == nil || *obj.ResetCounter != 2 {
		t.Fatalf("expected reset to be called twice, got %v", obj.ResetCounter)
	}
}

func TestPoolHandlesConcurrentAccess(t *testing.T) {
	p := New(newSample)
	const goroutines = 10

	done := make(chan struct{}, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			obj := p.Get()
			obj.Value = 99
			p.Put(obj)
			done <- struct{}{}
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}

	obj := p.Get()
	if obj.Value != 0 {
		t.Fatalf("expected reset value after concurrent puts, got %d", obj.Value)
	}
}

func TestPoolRetainsResetStateBetweenGets(t *testing.T) {
	p := New(newSample)
	obj := p.Get()
	obj.Value = 77
	p.Put(obj)

	reused := p.Get()
	reused.Value = 33
	p.Put(reused)

	if reused.ResetCounter == nil || *reused.ResetCounter != 2 {
		t.Fatalf("expected reset counter to increment on each put, got %v", reused.ResetCounter)
	}
}

func TestPoolPutNilPanics(t *testing.T) {
	p := New(newSample)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when putting nil object")
		}
	}()

	var obj *sample
	p.Put(obj)
}

func TestZeroValuePoolGetPanics(t *testing.T) {
	var p Pool[*sample]
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when getting from zero value pool")
		}
	}()

	_ = p.Get()
}

func TestPoolPutPropagatesResetPanic(t *testing.T) {
	p := New(func() *panicSample { return &panicSample{} })
	defer func() {
		if r := recover(); r != "reset failed" {
			t.Fatalf("expected panic %q, got %v", "reset failed", r)
		}
	}()

	p.Put(p.Get())
}

func BenchmarkPoolReuseHeavy(b *testing.B) {
	p := New(newHeavy)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		obj := p.Get()
		obj.data[0] = byte(i)
		p.Put(obj)
	}
}

func BenchmarkWithoutPoolHeavy(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		obj := newHeavy()
		obj.data[0] = byte(i)
		obj.Reset()
	}
}

func BenchmarkPoolReuseHeavyWithGCPressure(b *testing.B) {
	p := New(newHeavy)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		obj := p.Get()
		obj.data[len(obj.data)-1] = byte(i)
		if i%100 == 0 {
			runtime.GC()
		}
		p.Put(obj)
	}
}
