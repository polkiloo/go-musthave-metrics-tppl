package audit

import (
	"bufio"
	"context"
	"os"
	"testing"
)

func TestFileObserver_Notify_AppendsJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "audit-*.log")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer f.Close()

	observer := NewFileObserver(f.Name())
	event := Event{Timestamp: 42, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"}

	if err := observer.Notify(context.Background(), event); err != nil {
		t.Fatalf("notify: %v", err)
	}

	_ = f.Close()

	file, err := os.Open(f.Name())
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatalf("expected line in file")
	}
	got := scanner.Text()
	want := `{"ts":42,"metrics":["Alloc"],"ip_address":"127.0.0.1"}`
	if got != want {
		t.Fatalf("unexpected line: got %q want %q", got, want)
	}
}
