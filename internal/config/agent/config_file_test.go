package agentcfg

import (
	"path/filepath"
	"testing"
	"time"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
)

func TestParseDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "seconds string", input: "2s", want: 2 * time.Second},
		{name: "int seconds", input: "3", want: 3 * time.Second},
		{name: "invalid", input: "bad", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestReadAgentFileConfig(t *testing.T) {
	t.Parallel()

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()
		var cfg agentFileConfig
		if err := commoncfg.LoadConfigFile("", &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Address != nil || cfg.ReportInterval != nil {
			t.Fatalf("expected zero config for empty path, got %#v", cfg)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		var cfg agentFileConfig
		if err := commoncfg.LoadConfigFile("missing.json", &cfg); err == nil {
			t.Fatalf("expected error for missing file")
		}
	})
}

func TestBuildAgentConfig_FromTestdataFile(t *testing.T) {
	t.Parallel()

	cfgPath := filepath.Join("..", "..", "..", "testdata", "agent_config.json")
	withEnvMap(map[string]string{"CONFIG": cfgPath}, func() {
		withArgs(nil, func() {
			cfg, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Host != "file-agent" || cfg.Port != 12345 {
				t.Fatalf("want address file-agent:12345, got %s:%d", cfg.Host, cfg.Port)
			}
			if cfg.ReportInterval != 5*time.Second {
				t.Fatalf("want report interval 5s, got %v", cfg.ReportInterval)
			}
			if cfg.PollInterval != 2*time.Second {
				t.Fatalf("want poll interval 2s, got %v", cfg.PollInterval)
			}
			if cfg.SignKey != "agent-file-key" {
				t.Fatalf("want sign key from file, got %q", cfg.SignKey)
			}
			if cfg.RateLimit != 7 {
				t.Fatalf("want rate limit 7, got %d", cfg.RateLimit)
			}
			if cfg.CryptoKeyPath != "/path/to/agent.pem" {
				t.Fatalf("want crypto key path from file, got %q", cfg.CryptoKeyPath)
			}
		})
	})
}
