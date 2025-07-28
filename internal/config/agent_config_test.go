package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/config"
	"github.com/stretchr/testify/assert"
)

func writeTempFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)
	return filePath
}

func TestLoadConfig_ValidFullConfig(t *testing.T) {
	content := `
server_address: http://test-server:9000
poll_interval: 5s
report_interval: 20s
`
	path := writeTempFile(t, content)

	cfg, err := config.LoadAgentConfig(path)
	assert.NoError(t, err)
	assert.Equal(t, "http://test-server:9000", cfg.ServerAddress)
	assert.Equal(t, 5*time.Second, cfg.PollInterval)
	assert.Equal(t, 20*time.Second, cfg.ReportInterval)
}

func TestLoadConfig_MissingFields(t *testing.T) {
	content := `
poll_interval: 3s
`
	path := writeTempFile(t, content)

	cfg, err := config.LoadAgentConfig(path)
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", cfg.ServerAddress)
	assert.Equal(t, 3*time.Second, cfg.PollInterval)
	assert.Equal(t, 10*time.Second, cfg.ReportInterval)
}

func TestLoadConfig_FileNotExists(t *testing.T) {
	cfg, err := config.LoadAgentConfig("non-existent.yaml")
	assert.NoError(t, err)
	assert.Equal(t, config.DefaultConfig, cfg)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	content := `
  server_address: localhost
  poll_interval: !!notaduration
`
	path := writeTempFile(t, content)

	cfg, err := config.LoadAgentConfig(path)
	assert.Error(t, err)
	assert.Equal(t, config.DefaultConfig, cfg)
}
