package agent_test

import (
	"sync/atomic"
	"testing"
	"time"

	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"

	"github.com/stretchr/testify/assert"
)

type mockCollector struct{ collects int32 }

func (m *mockCollector) Collect() {
	atomic.AddInt32(&m.collects, 1)
}
func (m *mockCollector) Snapshot() (map[string]models.Gauge, map[string]models.Counter) {
	return map[string]models.Gauge{"Alloc": 1.23}, map[string]models.Counter{"PollCount": 2}
}

type mockSender struct{ sends int32 }

func (m *mockSender) Send(g map[string]models.Gauge, c map[string]models.Counter) {
	atomic.AddInt32(&m.sends, 1)
}

func TestAgentLoopSleep_Basic(t *testing.T) {
	c := &mockCollector{}
	s := &mockSender{}

	cfg := agent.AgentLoopConfig{
		PollInterval:   2 * time.Millisecond,
		ReportInterval: 5 * time.Millisecond,
		Iterations:     10,
	}

	agent.AgentLoopSleep(c, s, cfg)

	assert.GreaterOrEqual(t, atomic.LoadInt32(&c.collects), int32(10))
	assert.Greater(t, atomic.LoadInt32(&s.sends), int32(0))
}

func TestAgentLoopSleep_ZeroIterations(t *testing.T) {
	c := &mockCollector{}
	s := &mockSender{}
	done := make(chan struct{})

	go func() {
		cfg := agent.AgentLoopConfig{
			PollInterval:   1 * time.Millisecond,
			ReportInterval: 2 * time.Millisecond,
			Iterations:     0,
		}
		agent.AgentLoopSleep(c, s, cfg)
		close(done)
	}()
	select {
	case <-done:
		t.Error("should not finish when iterations=0")
	case <-time.After(10 * time.Millisecond):
	}
}

func TestAgentLoopSleep_ReportIntervalLongerThanLoop(t *testing.T) {
	c := &mockCollector{}
	s := &mockSender{}

	cfg := agent.AgentLoopConfig{
		PollInterval:   1 * time.Millisecond,
		ReportInterval: 100 * time.Millisecond,
		Iterations:     3,
	}
	agent.AgentLoopSleep(c, s, cfg)

	assert.Equal(t, int32(0), atomic.LoadInt32(&s.sends))
}
