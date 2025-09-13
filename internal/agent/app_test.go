package agent

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

type fakeLifecycle struct{ hooks []fx.Hook }

func (f *fakeLifecycle) Append(h fx.Hook) { f.hooks = append(f.hooks, h) }

func TestRunAgent_RegistersHooksAndStartsLoop_WithChan(t *testing.T) {
	collector := &test.FakeCollector{}
	s := &test.FakeAgentSenderWithChan{Ch: make(chan struct{}, 1)}
	cfg := AgentLoopConfig{
		PollInterval:   1 * time.Millisecond,
		ReportInterval: 2 * time.Millisecond,
		Iterations:     5,
		RateLimit:      1,
	}

	lc := &fakeLifecycle{}

	RunAgent(context.Background(), lc, collector, []sender.SenderInterface{s}, cfg)
	assert.Len(t, lc.hooks, 1)

	err := lc.hooks[0].OnStart(context.Background())
	assert.NoError(t, err)

	select {
	case <-s.Ch:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Send was not called in time")
	}

	assert.GreaterOrEqual(t, atomic.LoadInt32(&collector.Collected), int32(1))
	assert.GreaterOrEqual(t, atomic.LoadInt32(&s.Sends), int32(1))
	assert.NoError(t, lc.hooks[0].OnStop(context.Background()))
}

func TestProvideSender_ReturnsPlainAndJSON(t *testing.T) {
	cfg := AppConfig{Host: "localhost", Port: 8080}
	log := &test.FakeLogger{}
	comp := test.NewFakeCompressor("gzip")
	senders, err := ProvideSender(cfg, log, comp)
	if err != nil {
		t.Fatalf("ProvideSender returned error: %v", err)
	}
	if len(senders) != 2 {
		t.Fatalf("expected 2 senders, got %d", len(senders))
	}

	if gotType := reflect.TypeOf(senders[0]).String(); gotType != "*sender.PlainSender" {
		t.Errorf("expected first sender to be *sender.PlainSender, got %s", gotType)
	}
	if gotType := reflect.TypeOf(senders[1]).String(); gotType != "*sender.JSONSender" {
		t.Errorf("expected second sender to be *sender.JSONSender, got %s", gotType)
	}
}

func TestProvideAgentLoopConfig_CopiesFields(t *testing.T) {
	want := AppConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 5 * time.Second,
		LoopIterations: 10,
		RateLimit:      3,
	}

	got := ProvideAgentLoopConfig(want)

	if got.PollInterval != want.PollInterval {
		t.Errorf("PollInterval: want %v, got %v", want.PollInterval, got.PollInterval)
	}
	if got.ReportInterval != want.ReportInterval {
		t.Errorf("ReportInterval: want %v, got %v", want.ReportInterval, got.ReportInterval)
	}
	if got.Iterations != want.LoopIterations {
		t.Errorf("Iterations: want %d, got %d", want.LoopIterations, got.Iterations)
	}
	if got.RateLimit != want.RateLimit {
		t.Errorf("RateLimit: want %d, got %d", want.RateLimit, got.RateLimit)
	}
}

func TestProvideCollector_ReturnsCollector(t *testing.T) {
	cfg := AppConfig{}
	log := &test.FakeLogger{}

	c, err := ProvideCollector(cfg, log)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil collector, got nil")
	}
}

func TestRunAgent_AppContextCancellationStopsLoop(t *testing.T) {
	collector := &test.FakeCollector{}
	s := &test.FakeAgentSenderWithChan{Ch: make(chan struct{}, 1)}
	cfg := AgentLoopConfig{
		PollInterval:   1 * time.Millisecond,
		ReportInterval: 2 * time.Millisecond,
		Iterations:     0,
		RateLimit:      1,
	}

	appCtx, appCancel := context.WithCancel(context.Background())
	lc := &fakeLifecycle{}
	RunAgent(appCtx, lc, collector, []sender.SenderInterface{s}, cfg)
	assert.Len(t, lc.hooks, 1)

	err := lc.hooks[0].OnStart(context.Background())
	assert.NoError(t, err)

	select {
	case <-s.Ch:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Send was not called before cancel")
	}

	appCancel()

	select {
	case <-s.Ch:
		t.Fatal("unexpected Send after context cancel")
	case <-time.After(10 * time.Millisecond):
	}

	assert.NoError(t, lc.hooks[0].OnStop(context.Background()))
}
