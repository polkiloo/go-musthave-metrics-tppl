package main

import (
	"log"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/config"
)

func main() {
	cfg, err := config.LoadAgentConfig("config.yaml")
	if err != nil {
		log.Printf("config error: %v, using defaults", err)
	}

	collector := agent.NewCollector()
	sender := agent.NewSender("http://localhost", 8080)
	agent.AgentLoopSleep(collector, sender, cfg.PollInterval, cfg.ReportInterval, 0)

}
