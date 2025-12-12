package dbcfg

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"go.uber.org/fx"
)

func TestBuildDBConfig_EnvPrecedence(t *testing.T) {
	withEnv(EnvDSNVarName, "env", func() {
		withArgs([]string{"-d", "flag"}, func() {
			cfg, err := buildDBConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil || cfg.DSN != "env" {
				if cfg == nil {
					t.Fatalf("env should win, got nil config")
				}
				t.Fatalf("env should win, got %q", cfg.DSN)
			}
		})
	})
}

func TestBuildDBConfig_FlagWhenNoEnv(t *testing.T) {
	withEnv(EnvDSNVarName, "", func() {
		withArgs([]string{"-d", "flag"}, func() {
			cfg, err := buildDBConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil || cfg.DSN != "flag" {
				if cfg == nil {
					t.Fatalf("flag DSN expected, got nil config")
				}
				t.Fatalf("flag DSN expected, got %q", cfg.DSN)
			}
		})
	})
}

func TestModule_ProvidesConfig(t *testing.T) {
	withEnv(EnvDSNVarName, "envdsn", func() {
		var cfg *db.Config
		app := fx.New(
			fx.NopLogger,
			Module,
			fx.Invoke(func(c *db.Config) { cfg = c }),
		)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := app.Start(ctx); err != nil {
			t.Fatalf("fx.Start: %v", err)
		}
		defer func() { _ = app.Stop(ctx) }()
		if cfg == nil {
			t.Fatalf("config not provided")
		}
		if cfg.DSN != "envdsn" {
			t.Fatalf("want env DSN, got %q", cfg.DSN)
		}
	})
}

func TestBuildDBConfig_ConfigFileUsedWhenNoEnvOrFlag(t *testing.T) {
	cfgFile := t.TempDir() + "/config.json"
	if err := os.WriteFile(cfgFile, []byte(`{"database_dsn":"file-dsn"}`), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	withEnv("CONFIG", cfgFile, func() {
		withArgs(nil, func() {
			cfg, err := buildDBConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.DSN != "file-dsn" {
				t.Fatalf("expected DSN from file, got %q", cfg.DSN)
			}
		})
	})
}

func TestBuildDBConfig_EnvOverridesConfigFile(t *testing.T) {
	cfgFile := t.TempDir() + "/config.json"
	if err := os.WriteFile(cfgFile, []byte(`{"database_dsn":"file-dsn"}`), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	withEnv("CONFIG", cfgFile, func() {
		withEnv(EnvDSNVarName, "env-dsn", func() {
			cfg, err := buildDBConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.DSN != "env-dsn" {
				t.Fatalf("env should override file, got %q", cfg.DSN)
			}
		})
	})
}

func TestBuildDBConfig_FromTestdataFile(t *testing.T) {
	cfgPath := "../../../testdata/db_config.json"
	withEnv("CONFIG", cfgPath, func() {
		withArgs(nil, func() {
			cfg, err := buildDBConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.DSN != "dsn-from-testdata" {
				t.Fatalf("expected DSN from testdata, got %q", cfg.DSN)
			}
		})
	})
}

func withEnv(key, val string, fn func()) {
	old, had := os.LookupEnv(key)
	_ = os.Setenv(key, val)
	defer func() {
		if had {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	}()
	fn()
}

func withArgs(args []string, fn func()) {
	old := os.Args
	os.Args = append([]string{"cmd"}, args...)
	defer func() { os.Args = old }()
	fn()
}
