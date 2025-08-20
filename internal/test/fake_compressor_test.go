package test

import "github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"

var _ compression.Compressor = (*FakeCompressor)(nil)
