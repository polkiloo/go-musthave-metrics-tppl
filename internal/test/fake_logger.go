package test

import (
	"fmt"
	"sync"
)

type FakeLogger struct {
	mu     sync.Mutex
	infos  []string
	errors []string
}

func (f *FakeLogger) WriteInfo(msg string, kv ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fullMsg := msg
	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			fullMsg += fmt.Sprintf(" %v=%v", kv[i], kv[i+1])
		}
	}
	f.infos = append(f.infos, fullMsg)
}
func (f *FakeLogger) WriteError(msg string, kv ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fullMsg := msg
	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			fullMsg += fmt.Sprintf(" %v=%v", kv[i], kv[i+1])
		}
	}
	f.errors = append(f.errors, fullMsg)
}
func (f *FakeLogger) Sync() error { return nil }

func (f *FakeLogger) GetInfoMessages() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]string, len(f.infos))
	copy(result, f.infos)
	return result
}

func (f *FakeLogger) GetErrorMessages() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]string, len(f.errors))
	copy(result, f.errors)
	return result
}

func (f *FakeLogger) GetLastInfoMessage() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.infos) == 0 {
		return ""
	}
	return f.infos[len(f.infos)-1]
}
func (f *FakeLogger) GetLastErrorMessage() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.errors) == 0 {
		return ""
	}
	return f.errors[len(f.errors)-1]
}
