package main

import (
	"log"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
)

func main() {
	args, err := parseFlags()
	if err != nil {
		log.Fatalf("Error parsing: %v", err)
	}

	// cfg, err := config.LoadAgentConfig("config.yaml")
	// if err != nil {
	// 	log.Printf("config error: %v, using defaults", err)
	// }

	collector := agent.NewCollector()
	sender := agent.NewSender("http://"+args.Host, args.Port)
	agent.AgentLoopSleep(collector, sender, args.PollInterval, args.ReportInterval, 0)

}
