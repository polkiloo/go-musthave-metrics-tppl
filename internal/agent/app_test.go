package agent

import (
	"context"
	"testing"
	"time"

	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

type mockCollector struct{ collected int }

func (m *mockCollector) Collect() { m.collected++ }
func (m *mockCollector) Snapshot() (map[string]models.Gauge, map[string]models.Counter) {
	return map[string]models.Gauge{}, map[string]models.Counter{}
}

type mockSender struct{ sent int }

func (m *mockSender) Send(_ map[string]models.Gauge, _ map[string]models.Counter) { m.sent++ }

func TestProvideCollector(t *testing.T) {
	c := ProvideCollector()
	assert.NotNil(t, c)
}

func TestProvideSender(t *testing.T) {
	cfg := AppConfig{Host: "host", Port: 9999}
	s := ProvideSender(cfg)
	assert.NotNil(t, s)
}

func TestProvideConfig(t *testing.T) {
	cfg := AppConfig{
		ReportInterval: 5 * time.Second,
		PollInterval:   2 * time.Second,
	}
	loopCfg := ProvideConfig(cfg)
	assert.Equal(t, cfg.ReportInterval, loopCfg.ReportInterval)
	assert.Equal(t, cfg.PollInterval, loopCfg.PollInterval)
	assert.Equal(t, 0, loopCfg.Iterations)
}

type fakeLifecycle struct{ hooks []fx.Hook }

func (f *fakeLifecycle) Append(h fx.Hook) { f.hooks = append(f.hooks, h) }

type mockSenderWithChan struct {
	sent int
	ch   chan struct{}
}

func (m *mockSenderWithChan) Send(_ map[string]models.Gauge, _ map[string]models.Counter) {
	m.sent++
	select {
	case m.ch <- struct{}{}:
	default:
	}
}

func TestRunAgent_RegistersHooksAndStartsLoop_WithChan(t *testing.T) {
	collector := &mockCollector{}
	sender := &mockSenderWithChan{ch: make(chan struct{}, 1)}
	cfg := AgentLoopConfig{
		PollInterval:   1 * time.Millisecond,
		ReportInterval: 2 * time.Millisecond,
		Iterations:     5,
	}

	lc := &fakeLifecycle{}

	RunAgent(lc, collector, sender, cfg)
	assert.Len(t, lc.hooks, 1)

	err := lc.hooks[0].OnStart(context.Background())
	assert.NoError(t, err)

	select {
	case <-sender.ch:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Send was not called in time")
	}

	assert.GreaterOrEqual(t, collector.collected, 1)
	assert.GreaterOrEqual(t, sender.sent, 1)
	assert.NoError(t, lc.hooks[0].OnStop(context.Background()))
}
