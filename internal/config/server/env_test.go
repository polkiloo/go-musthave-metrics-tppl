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

func TestGetEnvVars_StoreInterval(t *testing.T) {
	withEnv(EnvStoreIntervalVarName, "10", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.StoreInterval == nil || *got.StoreInterval != 10 {
			t.Fatalf("want StoreInterval=10, got %v", got.StoreInterval)
		}
	})
}

func TestGetEnvVars_FileStorageAndRestore(t *testing.T) {
	withEnv(EnvFileStorageVarName, "/tmp/file.json", func() {
		withEnv(EnvRestoreVarName, "true", func() {
			got, err := getEnvVars()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.FileStorage != "/tmp/file.json" {
				t.Fatalf("file path mismatch: %q", got.FileStorage)
			}
			if got.Restore == nil || *got.Restore != true {
				t.Fatalf("want Restore=true, got %v", got.Restore)
			}
		})
	})
}

func TestGetEnvVars_InvalidStoreIntervalOrRestore(t *testing.T) {
	withEnv(EnvStoreIntervalVarName, "bad", func() {
		withEnv(EnvRestoreVarName, "badbool", func() {
			got, err := getEnvVars()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.StoreInterval != nil {
				t.Fatalf("want StoreInterval nil, got %v", *got.StoreInterval)
			}
			if got.Restore != nil {
				t.Fatalf("want Restore nil, got %v", *got.Restore)
			}
		})
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

func TestGetEnvVars_Key(t *testing.T) {
	withEnv(EnvKeyVarName, "secret", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SignKey != "secret" {
			t.Fatalf("key mismatch: %q", got.SignKey)
		}
	})
}

func TestGetEnvVars_Audit(t *testing.T) {
	withEnv(EnvAuditFileVarName, "/tmp/audit.log", func() {
		withEnv(EnvAuditURLVarName, "https://example.com/audit", func() {
			got, err := getEnvVars()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.AuditFile != "/tmp/audit.log" {
				t.Fatalf("audit file mismatch: %q", got.AuditFile)
			}
			if got.AuditURL != "https://example.com/audit" {
				t.Fatalf("audit url mismatch: %q", got.AuditURL)
			}
		})
	})
}

func TestGetEnvVars_TrustedSubnet(t *testing.T) {
	withEnv(EnvTrustedSubnet, "192.168.1.0/24", func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.TrustedSubnet != "192.168.1.0/24" {
			t.Fatalf("trusted subnet mismatch: %q", got.TrustedSubnet)
		}
	})
}
