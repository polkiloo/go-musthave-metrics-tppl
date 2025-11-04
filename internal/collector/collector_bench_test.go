package collector

import "testing"

func BenchmarkCollectorCollect(b *testing.B) {
	c := NewCollector()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Collect()
	}
}

func BenchmarkCollectorSnapshot(b *testing.B) {
	c := NewCollector()
	for i := 0; i < 1024; i++ {
		c.Collect()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Snapshot()
	}
}
