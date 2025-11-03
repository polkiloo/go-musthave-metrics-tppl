package test

import (
	"sync/atomic"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

type FakeAgentSender struct{ Sends int32 }

func (m *FakeAgentSender) Send(metrics []*models.Metrics) {
	atomic.AddInt32(&m.Sends, 1)
}
func (m *FakeAgentSender) SendBatch(metrics []*models.Metrics) { m.Send(metrics) }

type FakeAgentSenderWithChan struct {
	Sends int32
	Ch    chan struct{}
}

func (m *FakeAgentSenderWithChan) Send(metrics []*models.Metrics) {
	atomic.AddInt32(&m.Sends, 1)
	select {
	case m.Ch <- struct{}{}:
	default:
	}
}

func (m *FakeAgentSenderWithChan) SendBatch(metrics []*models.Metrics) {
	atomic.AddInt32(&m.Sends, 1)
	select {
	case m.Ch <- struct{}{}:
	default:
	}
}
