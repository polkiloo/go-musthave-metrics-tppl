package config_test

import (
	"os"
	"testing"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/config"
)

func TestLoadServerConfig_DefaultOnMissingFile(t *testing.T) {
	cfg, err := config.LoadServerConfig("nonexistent.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ServerPort != config.DefaultServerConfig.ServerPort {
		t.Errorf("expected default port %q, got %q", config.DefaultServerConfig.ServerPort, cfg.ServerPort)
	}
}

func TestLoadServerConfig_ValidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "server_config_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "server_port: \"9090\"\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := config.LoadServerConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("expected port \"9090\", got %q", cfg.ServerPort)
	}
}

func TestLoadServerConfig_InvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "server_config_invalid_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "server_port: [not a string]\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := config.LoadServerConfig(tmpFile.Name())
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
	if cfg.ServerPort != config.DefaultServerConfig.ServerPort {
		t.Errorf("expected default port %q on error, got %q", config.DefaultServerConfig.ServerPort, cfg.ServerPort)
	}
}
