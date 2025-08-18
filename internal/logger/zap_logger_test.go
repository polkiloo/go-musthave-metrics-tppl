package logger

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNew_Smoke(t *testing.T) {
	l, err := NewZapLogger()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if l == nil {
		t.Fatalf("New() returned nil logger")
	}
}

func newObservedZapLogger(level zap.AtomicLevel) (*ZapLogger, *observer.ObservedLogs) {
	core, logs := observer.New(level.Level())
	z := zap.New(core)
	return &ZapLogger{z.Sugar()}, logs
}

func TestWriteInfo_CapturesMessageAndFields(t *testing.T) {
	zl, logs := newObservedZapLogger(zap.NewAtomicLevelAt(zap.DebugLevel))

	zl.WriteInfo("hello", "k1", "v1", "k2", 42)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	ent := entries[0]
	if ent.Entry.Message != "hello" {
		t.Errorf("message = %q; want %q", ent.Entry.Message, "hello")
	}
	if ent.Entry.Level != zap.InfoLevel {
		t.Errorf("level = %v; want %v", ent.Entry.Level, zap.InfoLevel)
	}

	got := make(map[string]interface{})
	for _, field := range ent.Context {
		switch field.Type {
		case 15:
			got[field.Key] = field.String
		case 11:
			got[field.Key] = int(field.Integer)
		default:
			got[field.Key] = field.Interface
		}
	}

	t.Logf("Extracted fields: %+v", got)
	t.Logf("k1 value: %v (type: %T)", got["k1"], got["k1"])
	t.Logf("k2 value: %v (type: %T)", got["k2"], got["k2"])

	if got["k1"] != "v1" {
		t.Errorf("field k1 = %v; want %v", got["k1"], "v1")
	}
	if v, ok := got["k2"]; !ok || v != 42 {
		t.Errorf("field k2 = %v (ok=%v); want %v", v, ok, 42)
	}

	if err := zl.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

func TestWriteError_CapturesMessageAndFields(t *testing.T) {
	zl, logs := newObservedZapLogger(zap.NewAtomicLevelAt(zap.DebugLevel))

	zl.WriteError("oops", "code", 500, "detail", "boom")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	ent := entries[0]
	if ent.Entry.Message != "oops" {
		t.Errorf("message = %q; want %q", ent.Entry.Message, "oops")
	}
	if ent.Entry.Level != zap.ErrorLevel {
		t.Errorf("level = %v; want %v", ent.Entry.Level, zap.ErrorLevel)
	}

	got := make(map[string]interface{})
	for _, field := range ent.Context {
		switch field.Type {
		case 15:
			got[field.Key] = field.String
		case 11:
			got[field.Key] = int(field.Integer)
		default:
			got[field.Key] = field.Interface
		}
	}
	if got["code"] != 500 {
		t.Errorf("field code = %v; want %v", got["code"], 500)
	}
	if got["detail"] != "boom" {
		t.Errorf("field detail = %v; want %v", got["detail"], "boom")
	}
}

func TestNew_BuildError(t *testing.T) {
	orig := buildZapLogger
	defer func() { buildZapLogger = orig }()

	buildZapLogger = func(cfg zap.Config) (*zap.Logger, error) {
		return nil, errors.New("build failed")
	}

	l, err := NewZapLogger()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if l != nil {
		t.Fatalf("expected nil logger, got %#v", l)
	}
}

func TestWriteInfo_WithoutFields(t *testing.T) {
	zl, logs := newObservedZapLogger(zap.NewAtomicLevelAt(zap.DebugLevel))

	zl.WriteInfo("hello")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	ent := entries[0]
	if ent.Entry.Message != "hello" {
		t.Errorf("message = %q; want %q", ent.Entry.Message, "hello")
	}
	if ent.Entry.Level != zap.InfoLevel {
		t.Errorf("level = %v; want %v", ent.Entry.Level, zap.InfoLevel)
	}
}

func TestWriteError_WithoutFields(t *testing.T) {
	zl, logs := newObservedZapLogger(zap.NewAtomicLevelAt(zap.DebugLevel))

	zl.WriteError("oops")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	ent := entries[0]
	if ent.Entry.Message != "oops" {
		t.Errorf("message = %q; want %q", ent.Entry.Message, "oops")
	}
	if ent.Entry.Level != zap.ErrorLevel {
		t.Errorf("level = %v; want %v", ent.Entry.Level, zap.ErrorLevel)
	}
}
