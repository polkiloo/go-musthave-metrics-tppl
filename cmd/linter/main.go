package main

import (
	"github.com/polkiloo/go-musthave-metrics-tppl/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
