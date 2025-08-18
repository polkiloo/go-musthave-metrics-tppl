package models

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// func TestInitSets(t *testing.T) {
// 	t.Parallel()

// 	GaugeSet = map[string]struct{}{}
// 	CounterSet = map[string]struct{}{}
// 	if IsGaugeName("Alloc") || IsCounterName("PollCount") {
// 		t.Fatalf("sets must be empty before initSets")
// 	}

// 	initSets()

// 	if !IsGaugeName("Alloc") {
// 		t.Fatalf("gaugeSet not initialized with 'Alloc'")
// 	}
// 	if !IsCounterName("PollCount") {
// 		t.Fatalf("counterSet not initialized with 'PollCount'")
// 	}
// }

// func TestIsNameHelpers(t *testing.T) {
// 	t.Parallel()
// 	if !IsGaugeName("HeapAlloc") {
// 		t.Errorf("expected HeapAlloc to be gauge")
// 	}
// 	if IsGaugeName("PollCount") {
// 		t.Errorf("PollCount must not be gauge")
// 	}
// 	if !IsCounterName("PollCount") {
// 		t.Errorf("expected PollCount to be counter")
// 	}
// 	if IsCounterName("Alloc") {
// 		t.Errorf("Alloc must not be counter")
// 	}
// 	if IsGaugeName("___nope___") || IsCounterName("___nope___") {
// 		t.Errorf("unknown names must not match")
// 	}
// }

func TestNewGaugeMetrics(t *testing.T) {
	t.Parallel()

	val := 1.23
	m, err := NewGaugeMetrics("Alloc", &val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != "Alloc" || m.MType != GaugeType {
		t.Fatalf("unexpected metric fields: %+v", m)
	}
	if m.Value == nil || *m.Value != val {
		t.Fatalf("value not set properly")
	}
	if m.Delta != nil {
		t.Fatalf("delta must be nil for gauge")
	}

	m, err = NewGaugeMetrics("Alloc", nil)
	if err != nil {
		t.Fatalf("unexpected error with nil value: %v", err)
	}
	if m.Value != nil {
		t.Fatalf("expected nil Value")
	}

	// _, err = NewGaugeMetrics("PollCount", &val)
	// if !errors.Is(err, ErrMetricNameTypeMismatch) {
	// 	t.Fatalf("expected ErrMetricNameTypeMismatch, got %v", err)
	// }

	// _, err = NewGaugeMetrics("UnknownGauge", &val)
	// if !errors.Is(err, ErrMetricUnknownName) {
	// 	t.Fatalf("expected ErrMetricUnknownName, got %v", err)
	// }
}

func TestNewCounterMetrics(t *testing.T) {
	t.Parallel()

	val := int64(42)
	m, err := NewCounterMetrics("PollCount", &val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != "PollCount" || m.MType != CounterType {
		t.Fatalf("unexpected metric fields: %+v", m)
	}
	if m.Delta == nil || *m.Delta != val {
		t.Fatalf("delta not set properly")
	}
	if m.Value != nil {
		t.Fatalf("value must be nil for counter")
	}

	m, err = NewCounterMetrics("PollCount", nil)
	if err != nil {
		t.Fatalf("unexpected error with nil value: %v", err)
	}
	if m.Delta != nil {
		t.Fatalf("expected nil Delta")
	}

	// _, err = NewCounterMetrics("Alloc", &val)
	// if !errors.Is(err, ErrMetricNameTypeMismatch) {
	// 	t.Fatalf("expected ErrMetricNameTypeMismatch, got %v", err)
	// }

	// _, err = NewCounterMetrics("UnknownCounter", &val)
	// if !errors.Is(err, ErrMetricUnknownName) {
	// 	t.Fatalf("expected ErrUnknownMetricName, got %v", err)
	// }
}

func TestNewMetrics_FromString(t *testing.T) {
	t.Parallel()

	m, err := NewMetrics("Alloc", "3.14", GaugeType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.MType != GaugeType || m.Value == nil || *m.Value != 3.14 {
		t.Fatalf("unexpected metric: %+v", m)
	}
	if m.Delta != nil {
		t.Fatalf("delta must be nil for gauge")
	}

	m, err = NewMetrics("PollCount", "100", CounterType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.MType != CounterType || m.Delta == nil || *m.Delta != 100 {
		t.Fatalf("unexpected metric: %+v", m)
	}
	if m.Value != nil {
		t.Fatalf("value must be nil for counter")
	}

	m, err = NewMetrics("Alloc", "", GaugeType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Value != nil || m.MType != GaugeType {
		t.Fatalf("expected nil Value gauge: %+v", m)
	}

	m, err = NewMetrics("PollCount", "", CounterType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Delta != nil || m.MType != CounterType {
		t.Fatalf("expected nil Delta counter: %+v", m)
	}

	_, err = NewMetrics("Alloc", "not-a-float", GaugeType)
	if !errors.Is(err, ErrMetricInvalidValueType) {
		t.Fatalf("expected ErrMetricInvalidValueType, got %v", err)
	}

	_, err = NewMetrics("PollCount", "12.3", CounterType)
	if !errors.Is(err, ErrMetricInvalidValueType) {
		t.Fatalf("expected ErrMetricInvalidValueType, got %v", err)
	}

}

func TestMarshalJSON_Success(t *testing.T) {
	t.Parallel()

	val := 1.5
	m := &Metrics{ID: "Alloc", MType: GaugeType, Value: &val}
	b, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var back Metrics
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal back: %v", err)
	}
	if back.MType != GaugeType || back.ID != "Alloc" || back.Value == nil || *back.Value != val {
		t.Fatalf("roundtrip failed: %+v", back)
	}

	d := int64(7)
	m = &Metrics{ID: "PollCount", MType: CounterType, Delta: &d}
	b, err = m.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	back = Metrics{}
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal back: %v", err)
	}
	if back.MType != CounterType || back.ID != "PollCount" || back.Delta == nil || *back.Delta != d {
		t.Fatalf("roundtrip failed: %+v", back)
	}
}

// func TestMarshalJSON_Errors(t *testing.T) {
// 	t.Parallel()

// 	var nilM *Metrics
// 	b, err := nilM.MarshalJSON()
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
// 	if string(b) != "null" {
// 		t.Fatalf("expected null, got %s", string(b))
// 	}

// 	m := &Metrics{ID: "Alloc", MType: MetricType("weird")}
// 	_, err = m.MarshalJSON()
// 	if !errors.Is(err, ErrMetricInvalidType) {
// 		t.Fatalf("expected ErrMetricInvalidType, got %v", err)
// 	}

// 	m = &Metrics{ID: "Alloc", MType: GaugeType, Value: nil}
// 	_, err = m.MarshalJSON()
// 	if !errors.Is(err, ErrMetricMissingValue) {
// 		t.Fatalf("expected ErrMetricMissingValue, got %v", err)
// 	}

// 	m = &Metrics{ID: "PollCount", MType: CounterType, Delta: nil}
// 	_, err = m.MarshalJSON()
// 	if !errors.Is(err, ErrMetricMissingValue) {
// 		t.Fatalf("expected ErrMetricMissingValue, got %v", err)
// 	}

// 	val := 2.0
// 	d := int64(3)
// 	m = &Metrics{ID: "Alloc", MType: GaugeType, Value: &val, Delta: &d}
// 	_, err = m.MarshalJSON()
// 	if !errors.Is(err, ErrMetricAmbiguousValue) {
// 		t.Fatalf("expected ErrMetricAmbiguousValue, got %v", err)
// 	}

// 	m = &Metrics{ID: "PollCount", MType: CounterType, Delta: &d, Value: &val}
// 	_, err = m.MarshalJSON()
// 	if !errors.Is(err, ErrMetricAmbiguousValue) {
// 		t.Fatalf("expected ErrMetricAmbiguousValue, got %v", err)
// 	}

// }

func TestUnmarshalJSON_Success(t *testing.T) {
	t.Parallel()

	src := []byte(`{"id":"Alloc","type":"gauge","value":1.25}`)
	var m Metrics
	if err := m.UnmarshalJSON(src); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.MType != GaugeType || m.ID != "Alloc" || m.Value == nil || *m.Value != 1.25 || m.Delta != nil {
		t.Fatalf("unexpected m: %+v", m)
	}

	src = []byte(`{"id":"PollCount","type":"counter","delta":33}`)
	m = Metrics{}
	if err := m.UnmarshalJSON(src); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.MType != CounterType || m.ID != "PollCount" || m.Delta == nil || *m.Delta != 33 || m.Value != nil {
		t.Fatalf("unexpected m: %+v", m)
	}
}

func TestUnmarshalJSON_Errors(t *testing.T) {
	t.Parallel()

	var m Metrics
	if err := m.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Fatalf("expected syntax error")
	}

	// if err := m.UnmarshalJSON([]byte(`{"id":"Alloc","type":"weird"}`)); !errors.Is(err, ErrMetricInvalidType) {
	// 	t.Fatalf("expected ErrInvalidType, got %v", err)
	// }

	// if err := m.UnmarshalJSON([]byte(`{"id":"Alloc","type":"gauge"}`)); !errors.Is(err, ErrMetricMissingValue) {
	// 	t.Fatalf("expected ErrMetricMissingValue, got %v", err)
	// }

	// if err := m.UnmarshalJSON([]byte(`{"id":"PollCount","type":"counter"}`)); !errors.Is(err, ErrMetricMissingValue) {
	// 	t.Fatalf("expected ErrMetricMissingValue, got %v", err)
	// }

	// if err := m.UnmarshalJSON([]byte(`{"id":"Alloc","type":"gauge","value":1.0,"delta":1}`)); !errors.Is(err, ErrMetricAmbiguousValue) {
	// 	t.Fatalf("expected ErrMetricAmbiguousValue, got %v", err)
	// }

	// if err := m.UnmarshalJSON([]byte(`{"id":"PollCount","type":"counter","delta":1,"value":1.0}`)); !errors.Is(err, ErrMetricAmbiguousValue) {
	// 	t.Fatalf("expected ErrMetricAmbiguousValue, got %v", err)
	// }

	// if err := m.UnmarshalJSON([]byte(`{"id":"Alloc","type":"counter","delta":1}`)); !errors.Is(err, ErrMetricNameTypeMismatch) {
	// 	t.Fatalf("expected ErrMetricNameTypeMismatch, got %v", err)
	// }

	// if err := m.UnmarshalJSON([]byte(`{"id":"PollCount","type":"gauge","value":1.0}`)); !errors.Is(err, ErrMetricNameTypeMismatch) {
	// 	t.Fatalf("expected ErrMetricNameTypeMismatch, got %v", err)
	// }

	// if err := m.UnmarshalJSON([]byte(`{"id":"NoSuchGauge","type":"gauge","value":1.0}`)); !errors.Is(err, ErrMetricUnknownName) {
	// 	t.Fatalf("expected ErrUnknownMetricName, got %v", err)
	// }

	// if err := m.UnmarshalJSON([]byte(`{"id":"NoSuchCounter","type":"counter","delta":1}`)); !errors.Is(err, ErrMetricUnknownName) {
	// 	t.Fatalf("expected ErrUnknownMetricName, got %v", err)
	// }
}

func TestMetricTypes_OrderAndContents(t *testing.T) {
	t.Parallel()

	want := []MetricType{GaugeType, CounterType}
	if !reflect.DeepEqual(MetricTypes, want) {
		t.Fatalf("MetricTypes mismatch.\n got: %v\nwant: %v", MetricTypes, want)
	}
}

func TestMetricType_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		t    MetricType
		want bool
	}{
		{"gauge valid", GaugeType, true},
		{"counter valid", CounterType, true},
		{"unknown invalid", MetricType("weird"), false},
		{"empty invalid", MetricType(""), false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.t.IsValid(); got != tt.want {
				t.Fatalf("IsValid(%q) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

func TestParseMetricType(t *testing.T) {
	t.Parallel()

	mt, err := ParseMetricType("gauge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != GaugeType {
		t.Fatalf("ParseMetricType(gauge) = %v, want %v", mt, GaugeType)
	}

	mt, err = ParseMetricType("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != CounterType {
		t.Fatalf("ParseMetricType(counter) = %v, want %v", mt, CounterType)
	}

	_, err = ParseMetricType("nope")
	if !errors.Is(err, ErrMetricInvalidType) {
		t.Fatalf("expected ErrMetricInvalidType, got %v", err)
	}

	_, err = ParseMetricType("")
	if !errors.Is(err, ErrMetricInvalidType) {
		t.Fatalf("expected ErrMetricInvalidType for empty, got %v", err)
	}
}
