package test

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sender"
)

var _ sender.SenderInterface = &FakeAgentSender{}
var _ sender.SenderInterface = &FakeAgentSenderWithChan{}
