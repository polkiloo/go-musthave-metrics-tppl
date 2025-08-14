package servercfg

import (
	"os"
	"testing"
)

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

func TestGetEnvVars_NotSet(t *testing.T) {
	withEnv(EnvAddressVarName, "", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" {
			t.Fatalf("want empty host, got %q", got.Host)
		}
		if got.Port != nil {
			t.Fatalf("want nil Port, got %v", *got.Port)
		}
	})
}

func TestGetEnvVars_SetValid(t *testing.T) {
	withEnv(EnvAddressVarName, "example.com:8181", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "example.com" {
			t.Fatalf("want host=example.com, got %q", got.Host)
		}
		if got.Port == nil || *got.Port != 8181 {
			if got.Port == nil {
				t.Fatalf("want port=8181, got <nil>")
			}
			t.Fatalf("want port=8181, got %d", *got.Port)
		}
	})
}

func TestGetEnvVars_SetValid_EmptyHost(t *testing.T) {
	withEnv(EnvAddressVarName, ":6060", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" {
			t.Fatalf("want empty host, got %q", got.Host)
		}
		if got.Port == nil || *got.Port != 6060 {
			if got.Port == nil {
				t.Fatalf("want port=6060, got <nil>")
			}
			t.Fatalf("want port=6060, got %d", *got.Port)
		}
	})
}

func TestGetEnvVars_SetInvalid(t *testing.T) {
	withEnv(EnvAddressVarName, "not-an-addr", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" {
			t.Fatalf("want empty host, got %q", got.Host)
		}
		if got.Port != nil {
			t.Fatalf("want nil Port, got %v", *got.Port)
		}
	})
}

func TestGetEnvVars_SetIPv6Bracketed(t *testing.T) {
	withEnv(EnvAddressVarName, "[2001:db8::1]:9091", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "2001:db8::1" {
			t.Fatalf("want IPv6 host '2001:db8::1', got %q", got.Host)
		}
		if got.Port == nil || *got.Port != 9091 {
			if got.Port == nil {
				t.Fatalf("want port=9091, got <nil>")
			}
			t.Fatalf("want port=9091, got %d", *got.Port)
		}
	})
}
