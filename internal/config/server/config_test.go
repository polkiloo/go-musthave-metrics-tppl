package servercfg

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"go.uber.org/fx"
)

func TestBuildServerConfig_Default_WhenNoEnvNoFlags(t *testing.T) {
	withEnv("ADDRESS", "", func() {
		withArgs(nil, func() {
			cfg, err := buildServerConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Host != server.DefaultAppHost {
				t.Fatalf("want default host %q, got %q", server.DefaultAppHost, cfg.Host)
			}
		})
	})
}

func TestBuildServerConfig_EnvAddressWins(t *testing.T) {
	withEnv("ADDRESS", "env-host:8080", func() {
		withArgs([]string{"-a", "flag-host:9090"}, func() {
			cfg, err := buildServerConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Host != "env-host" {
				t.Fatalf("env ADDRESS must win; got host=%q", cfg.Host)
			}
		})
	})
}

func TestBuildServerConfig_FlagUsedWhenNoEnv(t *testing.T) {
	withEnv("ADDRESS", "", func() {
		withArgs([]string{"-a", "flag-host:9090"}, func() {
			cfg, err := buildServerConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Host != "flag-host" {
				t.Fatalf("want flag host %q, got %q", "flag-host", cfg.Host)
			}
		})
	})
}

func TestBuildServerConfig_EmptyHostFromEnvAllowed(t *testing.T) {
	withEnv("ADDRESS", ":8081", func() {
		withArgs(nil, func() {
			cfg, err := buildServerConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Host != "localhost" {
				t.Fatalf("want empty host from ADDRESS ':8081', got %q", cfg.Host)
			}
		})
	})
}

func TestModule_AdapterValueToPointer_Executes(t *testing.T) {
	var gotVal server.AppConfig
	var gotPtr *server.AppConfig

	withEnv("ADDRESS", "", func() {
		withArgs(nil, func() {
			app := fx.New(
				fx.NopLogger,
				Module,
				fx.Invoke(func(v server.AppConfig, p *server.AppConfig) {
					gotVal = v
					gotPtr = p
				}),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := app.Start(ctx); err != nil {
				t.Fatalf("fx.Start: %v", err)
			}
			defer func() { _ = app.Stop(ctx) }()

			if gotPtr == nil {
				t.Fatalf("adapter not executed: *server.AppConfig is nil")
			}
			if !reflect.DeepEqual(*gotPtr, gotVal) {
				t.Fatalf("pointer content mismatch:\n got:  %+v\n want: %+v", *gotPtr, gotVal)
			}
		})
	})
}

func TestBuildServerConfig_KeyPriority(t *testing.T) {
	withEnv(EnvKeyVarName, "envkey", func() {
		withArgs([]string{"-k", "flagkey"}, func() {
			cfg, err := buildServerConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.SignKey != "envkey" {
				t.Fatalf("env key must win: got %q", cfg.SignKey)
			}
		})
	})
}

func TestBuildServerConfig_AuditPriority(t *testing.T) {
	withEnv(EnvAuditFileVarName, "/tmp/env.log", func() {
		withEnv(EnvAuditURLVarName, "https://env.example", func() {
			args := []string{"--audit-file", "/tmp/flag.log", "--audit-url", "https://flag.example"}
			withArgs(args, func() {
				cfg, err := buildServerConfig()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.AuditFile != "/tmp/env.log" {
					t.Fatalf("env audit file must win: got %q", cfg.AuditFile)
				}
				if cfg.AuditURL != "https://env.example" {
					t.Fatalf("env audit url must win: got %q", cfg.AuditURL)
				}
			})
		})
	})
}

func TestBuildServerConfig_ConfigFileAppliedBeforeEnvAndFlags(t *testing.T) {
	tmpFile := t.TempDir() + "/config.json"
	if err := os.WriteFile(tmpFile, []byte(`{
                "address": "file-host:9999",
                "store_interval": "5s",
                "store_file": "/tmp/file.db",
                "restore": false,
                "key": "filekey",
                "audit_file": "/tmp/audit.log",
                "audit_url": "https://file.example",
                "crypto_key": "/tmp/key.pem"
        }`), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	withEnv("CONFIG", tmpFile, func() {
		withEnv(EnvAddressVarName, "", func() {
			withArgs(nil, func() {
				cfg, err := buildServerConfig()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.Host != "file-host" || cfg.Port != 9999 {
					t.Fatalf("want address from file file-host:9999, got %s:%d", cfg.Host, cfg.Port)
				}
				if cfg.StoreInterval != 5 {
					t.Fatalf("want store interval 5 from file, got %d", cfg.StoreInterval)
				}
				if cfg.FileStoragePath != "/tmp/file.db" {
					t.Fatalf("want store file from file, got %q", cfg.FileStoragePath)
				}
				if cfg.Restore != false {
					t.Fatalf("want restore=false from file, got %v", cfg.Restore)
				}
				if cfg.SignKey != "filekey" {
					t.Fatalf("want key from file, got %q", cfg.SignKey)
				}
				if cfg.AuditFile != "/tmp/audit.log" {
					t.Fatalf("want audit file from file, got %q", cfg.AuditFile)
				}
				if cfg.AuditURL != "https://file.example" {
					t.Fatalf("want audit url from file, got %q", cfg.AuditURL)
				}
				if cfg.CryptoKeyPath != "/tmp/key.pem" {
					t.Fatalf("want crypto key from file, got %q", cfg.CryptoKeyPath)
				}
			})
		})
	})
}

func TestBuildServerConfig_EnvOverridesConfigFile(t *testing.T) {
	tmpFile := t.TempDir() + "/config.json"
	if err := os.WriteFile(tmpFile, []byte(`{"address": "file-host:9999", "store_interval": "15s"}`), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	withEnv("CONFIG", tmpFile, func() {
		withEnv(EnvAddressVarName, "env-host:8081", func() {
			withEnv(EnvStoreIntervalVarName, strconv.Itoa(20), func() {
				withArgs([]string{"-a", "flag-host:7777"}, func() {
					cfg, err := buildServerConfig()
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if cfg.Host != "env-host" || cfg.Port != 8081 {
						t.Fatalf("env should override config file, got %s:%d", cfg.Host, cfg.Port)
					}
					if cfg.StoreInterval != 20 {
						t.Fatalf("env store interval should override file, got %d", cfg.StoreInterval)
					}
				})
			})
		})
	})
}
