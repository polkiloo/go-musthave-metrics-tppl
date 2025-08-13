package config

import (
	"os"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	PollInterval   time.Duration `yaml:"poll_interval"`
	ReportInterval time.Duration `yaml:"report_interval"`
}

var DefaultConfig = AgentConfig{
	Host:           agent.DefaultAppConfig.Host,
	Port:           agent.DefaultAppConfig.Port,
	PollInterval:   agent.DefaultAppConfig.PollInterval,
	ReportInterval: agent.DefaultAppConfig.ReportInterval,
}

func LoadAgentConfig(path string) (AgentConfig, error) {
	cfg := DefaultConfig

	file, err := os.Open(path)
	if err != nil {
		return cfg, nil
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return DefaultConfig, err
	}

	return cfg, nil
}
