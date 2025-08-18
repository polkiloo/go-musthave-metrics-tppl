package test

import "github.com/polkiloo/go-musthave-metrics-tppl/internal/collector"

var _ collector.CollectorInterface = &FakeCollector{}
