package agent_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestAgentLoopSleep_Basic(t *testing.T) {
	c := &test.FakeCollector{}
	s := &test.FakeAgentSender{}

	cfg := agent.AgentLoopConfig{
		PollInterval:   2 * time.Millisecond,
		ReportInterval: 5 * time.Millisecond,
		Iterations:     10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	agent.AgentLoopSleep(ctx, c, []sender.SenderInterface{s}, cfg)

	assert.GreaterOrEqual(t, atomic.LoadInt32(&c.Collected), int32(10))
	assert.Greater(t, atomic.LoadInt32(&s.Sends), int32(0))
}

func TestAgentLoopSleep_ZeroIterations(t *testing.T) {
	c := &test.FakeCollector{}
	s := &test.FakeAgentSender{}
	done := make(chan struct{})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() {
		cfg := agent.AgentLoopConfig{
			PollInterval:   1 * time.Millisecond,
			ReportInterval: 2 * time.Millisecond,
			Iterations:     0,
		}
		agent.AgentLoopSleep(ctx, c, []sender.SenderInterface{s}, cfg)
		close(done)
	}()
	select {
	case <-done:
		t.Error("should not finish when iterations=0")
	case <-time.After(10 * time.Millisecond):
	}
}

func TestAgentLoopSleep_ReportIntervalLongerThanLoop(t *testing.T) {
	c := &test.FakeCollector{}
	s := &test.FakeAgentSender{}

	cfg := agent.AgentLoopConfig{
		PollInterval:   1 * time.Millisecond,
		ReportInterval: 100 * time.Millisecond,
		Iterations:     3,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	agent.AgentLoopSleep(ctx, c, []sender.SenderInterface{s}, cfg)

	assert.Equal(t, int32(0), atomic.LoadInt32(&s.Sends))
}
