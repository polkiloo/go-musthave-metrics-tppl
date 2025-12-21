package servercfg

import (
	"path/filepath"
	"testing"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
)

func TestParseDurationSeconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "duration format", input: "3s", want: 3},
		{name: "plain seconds", input: "4", want: 4},
		{name: "empty", input: "", wantErr: true},
		{name: "invalid", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseDurationSeconds(tt.input)
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
				t.Fatalf("want %d, got %d", tt.want, got)
			}
		})
	}
}

func TestReadServerFileConfig(t *testing.T) {
	t.Parallel()

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()
		var cfg serverFileConfig
		if err := commoncfg.LoadConfigFile("", &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Address != nil || cfg.Restore != nil {
			t.Fatalf("expected zero config, got %#v", cfg)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		var cfg serverFileConfig
		if err := commoncfg.LoadConfigFile("missing.json", &cfg); err == nil {
			t.Fatalf("expected error for missing file")
		}
	})
}

func TestBuildServerConfig_FromTestdataFile(t *testing.T) {
	t.Parallel()

	cfgPath := filepath.Join("..", "..", "..", "testdata", "server_config.json")
	withEnv("CONFIG", cfgPath, func() {
		withArgs(nil, func() {
			cfg, err := buildServerConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Host != "file-server" || cfg.Port != 9000 {
				t.Fatalf("want address file-server:9000, got %s:%d", cfg.Host, cfg.Port)
			}
			if cfg.StoreInterval != 12 {
				t.Fatalf("want store interval 12, got %d", cfg.StoreInterval)
			}
			if cfg.FileStoragePath != "/tmp/server.db" {
				t.Fatalf("want file storage from file, got %q", cfg.FileStoragePath)
			}
			if !cfg.Restore {
				t.Fatalf("want restore true from file")
			}
			if cfg.SignKey != "server-file-key" {
				t.Fatalf("want sign key from file, got %q", cfg.SignKey)
			}
			if cfg.AuditFile != "audit.log" {
				t.Fatalf("want audit file from file, got %q", cfg.AuditFile)
			}
			if cfg.AuditURL != "http://audit.local" {
				t.Fatalf("want audit url from file, got %q", cfg.AuditURL)
			}
			if cfg.CryptoKeyPath != "/path/to/server.pem" {
				t.Fatalf("want crypto key path from file, got %q", cfg.CryptoKeyPath)
			}
			if cfg.TrustedSubnet != "10.0.0.0/24" {
				t.Fatalf("want trusted subnet from file, got %q", cfg.TrustedSubnet)
			}
		})
	})
}
