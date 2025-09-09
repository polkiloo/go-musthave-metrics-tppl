package test

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

type FakeAgentSender struct{ Sends int32 }

func (m *FakeAgentSender) Send(metrics []*models.Metrics) { m.Sends++ }

func (m *FakeAgentSender) SendBatch(metrics []*models.Metrics) { m.Send(metrics) }

type FakeAgentSenderWithChan struct {
	Sends int
	Ch    chan struct{}
}

func (m *FakeAgentSenderWithChan) Send(metrics []*models.Metrics) {
	m.Sends++
	select {
	case m.Ch <- struct{}{}:
	default:
	}
}

func (m *FakeAgentSenderWithChan) SendBatch(metrics []*models.Metrics) {
	m.Sends++
	select {
	case m.Ch <- struct{}{}:
	default:
	}
}
