package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	ServerPort string `yaml:"server_port"`
}

var DefaultServerConfig = ServerConfig{
	ServerPort: "8080",
}

func LoadServerConfig(path string) (ServerConfig, error) {
	cfg := DefaultServerConfig

	file, err := os.Open(path)
	if err != nil {
		return cfg, nil
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return DefaultServerConfig, err
	}

	return cfg, nil
}
