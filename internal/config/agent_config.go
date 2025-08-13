package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	ServerAddress  string        `yaml:"server_address"`
	PollInterval   time.Duration `yaml:"poll_interval"`
	ReportInterval time.Duration `yaml:"report_interval"`
}

var DefaultConfig = AgentConfig{
	ServerAddress:  "http://localhost:8080",
	PollInterval:   2 * time.Second,
	ReportInterval: 10 * time.Second,
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
