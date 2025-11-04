package servercfg

import (
	"context"
	"reflect"
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
