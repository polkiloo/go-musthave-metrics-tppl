package collector

import (
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

func findMetric(snapshot []*models.Metrics, id string) (*models.Metrics, bool) {
	for _, m := range snapshot {
		if m != nil && m.ID == id {
			return m, true
		}
	}
	return nil, false
}

func TestCollector_Collect_PopulatesAllMetrics(t *testing.T) {

	c := NewCollector()
	c.Collect()

	snap := c.Snapshot()

	wantCount := len(gaugeGetters) + 1 + len(counterGetters)
	if len(snap) != wantCount {
		t.Fatalf("snapshot size = %d, want %d", len(snap), wantCount)
	}

	for name := range gaugeGetters {
		m, ok := findMetric(snap, name)
		if !ok {
			t.Fatalf("gauge %q not found in snapshot", name)
		}
		if m.MType != models.GaugeType {
			t.Fatalf("gauge %q has wrong type: %v", name, m.MType)
		}
		if m.Value == nil {
			t.Fatalf("gauge %q has nil Value", name)
		}
		if m.Delta != nil {
			t.Fatalf("gauge %q must have nil Delta", name)
		}
	}

	if rv, ok := findMetric(snap, "RandomValue"); !ok {
		t.Fatalf("RandomValue metric not found")
	} else {
		if rv.MType != models.GaugeType || rv.Value == nil {
			t.Fatalf("RandomValue type/value invalid: %+v", rv)
		}
		if *rv.Value < 0 || *rv.Value >= 100 {
			t.Fatalf("RandomValue out of range [0,100): %v", *rv.Value)
		}
	}

	for name := range counterGetters {
		m, ok := findMetric(snap, name)
		if !ok {
			t.Fatalf("counter %q not found in snapshot", name)
		}
		if m.MType != models.CounterType {
			t.Fatalf("counter %q has wrong type: %v", name, m.MType)
		}
		if m.Delta == nil {
			t.Fatalf("counter %q has nil Delta", name)
		}
		if m.Value != nil {
			t.Fatalf("counter %q must have nil Value", name)
		}
	}
}

func TestCollector_Counter_AccumulatesBetweenCollects(t *testing.T) {

	c := NewCollector()
	c.Collect()
	snap1 := c.Snapshot()

	m1, ok := findMetric(snap1, "PollCount")
	if !ok {
		t.Fatalf("PollCount not found after first Collect")
	}
	if m1.Delta == nil || *m1.Delta != 1 {
		t.Fatalf("PollCount after first Collect = %v, want 1", m1.Delta)
	}

	c.Collect()
	snap2 := c.Snapshot()

	m2, ok := findMetric(snap2, "PollCount")
	if !ok {
		t.Fatalf("PollCount not found after second Collect")
	}
	if m2.Delta == nil || *m2.Delta != 2 {
		t.Fatalf("PollCount after second Collect = %v, want 2", m2.Delta)
	}
}

func TestCollector_Snapshot_DeepCopy(t *testing.T) {

	c := NewCollector()
	c.Collect()

	snap1 := c.Snapshot()

	alloc1, ok := findMetric(snap1, "Alloc")
	if !ok {
		t.Fatalf("Alloc not found in snapshot")
	}
	origAlloc := *alloc1.Value

	pc1, ok := findMetric(snap1, "PollCount")
	if !ok {
		t.Fatalf("PollCount not found in snapshot")
	}
	origPoll := *pc1.Delta

	if alloc1.Value == nil || pc1.Delta == nil {
		t.Fatalf("unexpected nil pointers: alloc1=%v poll1=%v", alloc1.Value, pc1.Delta)
	}
	*alloc1.Value = 123456.0
	*pc1.Delta = 999

	snap2 := c.Snapshot()

	alloc2, ok := findMetric(snap2, "Alloc")
	if !ok || alloc2.Value == nil {
		t.Fatalf("Alloc not found or nil in second snapshot")
	}
	if *alloc2.Value != origAlloc {
		t.Fatalf("Alloc value leaked from snapshot mutation: got=%v want=%v", *alloc2.Value, origAlloc)
	}

	pc2, ok := findMetric(snap2, "PollCount")
	if !ok || pc2.Delta == nil {
		t.Fatalf("PollCount not found or nil in second snapshot")
	}
	if *pc2.Delta != origPoll {
		t.Fatalf("PollCount value leaked from snapshot mutation: got=%v want=%v", *pc2.Delta, origPoll)
	}
}

func TestCollector_Snapshot_ContainsAllAfterMultipleCollects(t *testing.T) {

	c := NewCollector()
	c.Collect()
	c.Collect()
	c.Collect()

	snap := c.Snapshot()

	wantCount := len(gaugeGetters) + 1 + len(counterGetters)
	if len(snap) != wantCount {
		t.Fatalf("snapshot size = %d, want %d", len(snap), wantCount)
	}

	checkIDs := []string{"Alloc", "HeapAlloc", "RandomValue", "PollCount"}
	for _, id := range checkIDs {
		if _, ok := findMetric(snap, id); !ok {
			t.Fatalf("metric %q missing after multiple collects", id)
		}
	}
}

func TestCloneMetrics_NilInput(t *testing.T) {
	if got := cloneMetrics(nil); got != nil {
		t.Errorf("expected nil, got %#v", got)
	}
}
