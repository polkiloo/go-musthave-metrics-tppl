package test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
)

func TestFakeLogger_WriteInfo_And_Getters(t *testing.T) {
	fl := &FakeLogger{}

	if got := fl.GetLastInfoMessage(); got != "" {
		t.Fatalf("GetLastInfoMessage on empty: %q; want \"\"", got)
	}
	if got := fl.GetLastErrorMessage(); got != "" {
		t.Fatalf("GetLastErrorMessage on empty: %q; want \"\"", got)
	}

	fl.WriteInfo("hello", "k1", "v1", "k2", 42, "orphan")
	infos := fl.GetInfoMessages()
	if len(infos) != 1 {
		t.Fatalf("infos len=%d; want 1", len(infos))
	}
	want1 := "hello k1=v1 k2=42"
	if infos[0] != want1 {
		t.Fatalf("info[0]=%q; want %q", infos[0], want1)
	}
	if last := fl.GetLastInfoMessage(); last != want1 {
		t.Fatalf("GetLastInfoMessage=%q; want %q", last, want1)
	}

	fl.WriteError("oops", "code", 500, "detail", "boom")
	errs := fl.GetErrorMessages()
	if len(errs) != 1 {
		t.Fatalf("errors len=%d; want 1", len(errs))
	}
	wantErr1 := "oops code=500 detail=boom"
	if errs[0] != wantErr1 {
		t.Fatalf("errors[0]=%q; want %q", errs[0], wantErr1)
	}
	if last := fl.GetLastErrorMessage(); last != wantErr1 {
		t.Fatalf("GetLastErrorMessage=%q; want %q", last, wantErr1)
	}

	infos[0] = "mutated"
	errs[0] = "mutated"
	if fl.GetLastInfoMessage() != want1 {
		t.Fatalf("internal infos mutated via copy")
	}
	if fl.GetLastErrorMessage() != wantErr1 {
		t.Fatalf("internal errors mutated via copy")
	}

	if err := fl.Sync(); err != nil {
		t.Fatalf("Sync() error: %v", err)
	}
}

func TestFakeLogger_Concurrency(t *testing.T) {
	fl := &FakeLogger{}
	const workers = 16
	const perWorker = 50

	var wg sync.WaitGroup
	wg.Add(workers * 2)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				fl.WriteInfo("info", "wid", id, "seq", j)
			}
		}(i)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				fl.WriteError("err", "wid", id, "seq", j)
			}
		}(i)
	}
	wg.Wait()

	infos := fl.GetInfoMessages()
	errs := fl.GetErrorMessages()

	if len(infos) != workers*perWorker {
		t.Fatalf("infos len=%d; want %d", len(infos), workers*perWorker)
	}
	if len(errs) != workers*perWorker {
		t.Fatalf("errors len=%d; want %d", len(errs), workers*perWorker)
	}

	lastInfo := fl.GetLastInfoMessage()
	lastErr := fl.GetLastErrorMessage()
	if lastInfo == "" || lastErr == "" {
		t.Fatalf("last messages should not be empty")
	}
}

func TestFakeLogger_WriteFormattingTypes(t *testing.T) {
	fl := &FakeLogger{}

	fl.WriteInfo("mixed", "int", 10, "str", "s", "bool", true)
	got := fl.GetLastInfoMessage()

	wantParts := []string{"mixed", "int=10", "str=s", "bool=true"}
	for _, part := range wantParts {
		if !contains(got, part) {
			t.Fatalf("want part %q in %q", part, got)
		}
	}

	fl.WriteError("many")
	for i := 0; i < 20; i++ {
		fl.WriteError("pairs", "k"+strconv.Itoa(i), i)
	}
	errs := fl.GetErrorMessages()
	if len(errs) < 21 {
		t.Fatalf("errors len=%d; want >=21", len(errs))
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
outer:
	for i := 0; i+len(sub) <= len(s); i++ {
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}

var _ logger.Logger = &FakeLogger{}
