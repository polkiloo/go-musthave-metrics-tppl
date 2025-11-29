package servercfg

import (
	"os"
	"testing"
)

func withArgs(args []string, fn func()) {
	old := os.Args
	os.Args = append([]string{"cmd"}, args...)
	defer func() { os.Args = old }()
	fn()
}

func TestParseFlags_NoArgs(t *testing.T) {
	withArgs(nil, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.addressFlag.Host != "" {
			t.Fatalf("want empty host (flag not set), got %q", got.addressFlag.Host)
		}
		if got.addressFlag.Port != nil {
			t.Fatalf("want nil port (flag not set), got %v", *got.addressFlag.Port)
		}
	})
}

func TestParseFlags_AInline(t *testing.T) {
	withArgs([]string{"-a=1.2.3.4:9999"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.addressFlag.Host != "1.2.3.4" {
			t.Fatalf("host mismatch: %q", got.addressFlag.Host)
		}
		if got.addressFlag.Port == nil || *got.addressFlag.Port != 9999 {
			if got.addressFlag.Port == nil {
				t.Fatalf("port nil; want 9999")
			}
			t.Fatalf("port mismatch: %d", *got.addressFlag.Port)
		}
	})
}

func TestParseFlags_ASeparate(t *testing.T) {
	withArgs([]string{"-a", "localhost:5555"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.addressFlag.Host != "localhost" {
			t.Fatalf("host mismatch: %q", got.addressFlag.Host)
		}
		if got.addressFlag.Port == nil || *got.addressFlag.Port != 5555 {
			t.Fatalf("port mismatch: %v", got.addressFlag.Port)
		}
	})
}

func TestParseFlags_EmptyHost(t *testing.T) {
	withArgs([]string{"-a", ":7777"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.addressFlag.Host != "" {
			t.Fatalf("empty host must remain empty, got %q", got.addressFlag.Host)
		}
		if got.addressFlag.Port == nil || *got.addressFlag.Port != 7777 {
			t.Fatalf("port mismatch: %v", got.addressFlag.Port)
		}
	})
}

func TestParseFlags_IPv6Bracketed(t *testing.T) {
	withArgs([]string{"-a", "[::1]:8081"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.addressFlag.Host != "::1" {
			t.Fatalf("IPv6 host mismatch: %q", got.addressFlag.Host)
		}
		if got.addressFlag.Port == nil || *got.addressFlag.Port != 8081 {
			t.Fatalf("port mismatch: %v", got.addressFlag.Port)
		}
	})
}

func TestParseFlags_InvalidA_ReturnsError(t *testing.T) {
	withArgs([]string{"-a", "not-an-addr"}, func() {
		_, err := parseFlags()
		if err == nil {
			t.Fatalf("expected error for invalid -a")
		}
	})
}

func TestParseFlags_UnknownFlag_Error(t *testing.T) {
	withArgs([]string{"-x"}, func() {
		_, err := parseFlags()
		if err == nil {
			t.Fatalf("expected error for unknown flag")
		}
	})
}

func TestParseFlags_StoreInterval(t *testing.T) {
	withArgs([]string{"-i", "42"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.storeInterval == nil || *got.storeInterval != 42 {
			t.Fatalf("store interval mismatch: %v", got.storeInterval)
		}
	})
}
func TestParseFlags_FileStorageAndRestore(t *testing.T) {
	withArgs([]string{"-f", "/tmp/file.json", "-r=true"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.fileStorage != "/tmp/file.json" {
			t.Fatalf("file storage mismatch: %q", got.fileStorage)
		}
		if got.restore == nil || *got.restore != true {
			t.Fatalf("restore mismatch: %v", got.restore)
		}
	})
}

func TestParseFlags_InvalidValues(t *testing.T) {
	withArgs([]string{"-i", "bad"}, func() {
		if _, err := parseFlags(); err == nil {
			t.Fatalf("expected error for invalid -i")
		}
	})
	withArgs([]string{"-r=badbool"}, func() {
		if _, err := parseFlags(); err == nil {
			t.Fatalf("expected error for invalid -r")
		}
	})
}

func TestParseFlags_Key(t *testing.T) {
	withArgs([]string{"-k", "secret"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SignKey != "secret" {
			t.Fatalf("key mismatch: %q", got.SignKey)
		}
	})
}

func TestParseFlags_AuditFlags(t *testing.T) {
	withArgs([]string{"--audit-file", "/tmp/audit.log", "--audit-url", "https://example.com/audit"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.auditFile != "/tmp/audit.log" {
			t.Fatalf("audit file mismatch: %q", got.auditFile)
		}
		if got.auditURL != "https://example.com/audit" {
			t.Fatalf("audit url mismatch: %q", got.auditURL)
		}
	})
}
