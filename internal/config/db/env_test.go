package dbcfg

import (
	"testing"
)

func TestGetEnvVars_NotSet(t *testing.T) {
	withEnv(EnvDSNVarName, "", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.DSN != "" {
			t.Fatalf("want empty DSN, got %q", got.DSN)
		}
	})
}

func TestGetEnvVars_Set(t *testing.T) {
	withEnv(EnvDSNVarName, "postgres://example.com/db", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.DSN != "postgres://example.com/db" {
			t.Fatalf("want DSN 'postgres://example.com/db', got %q", got.DSN)
		}
	})
}
