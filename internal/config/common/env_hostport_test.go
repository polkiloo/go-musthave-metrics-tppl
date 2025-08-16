package commoncfg

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

func TestSplitHostPort_ValidIPv4(t *testing.T) {
	h, p, ok := splitHostPort("localhost:8080")
	if !ok || h != "localhost" || p != 8080 {
		t.Fatalf("want localhost:8080, got %q:%d (ok=%v)", h, p, ok)
	}
}

func TestSplitHostPort_EmptyHost_AllIfaces(t *testing.T) {
	h, p, ok := splitHostPort(":9090")
	if !ok || h != "" || p != 9090 {
		t.Fatalf("want '':9090, got %q:%d (ok=%v)", h, p, ok)
	}
}

func TestSplitHostPort_IPv6Bracketed(t *testing.T) {
	h, p, ok := splitHostPort("[::1]:7070")
	if !ok || h != "::1" || p != 7070 {
		t.Fatalf("want [::1]:7070, got %q:%d (ok=%v)", h, p, ok)
	}
}

func TestSplitHostPort_Invalid_NoPort(t *testing.T) {
	if _, _, ok := splitHostPort("localhost"); ok {
		t.Fatalf("expected not ok (no port)")
	}
}

func TestSplitHostPort_Invalid_ZeroPort(t *testing.T) {
	if _, _, ok := splitHostPort("127.0.0.1:0"); ok {
		t.Fatalf("expected not ok (zero port)")
	}
}

func TestSplitHostPort_Invalid_NonNumeric(t *testing.T) {
	if _, _, ok := splitHostPort("127.0.0.1:http"); ok {
		t.Fatalf("expected not ok (non numeric)")
	}
}

func TestSplitHostPort_Invalid_IPv6WithoutBrackets(t *testing.T) {
	if _, _, ok := splitHostPort("::1:8080"); ok {
		t.Fatalf("expected not ok (ipv6 without brackets)")
	}
}

func TestReadHostPortEnv_NotSet(t *testing.T) {
	withEnv("ADDRESS", "", func() {
		got, err := ReadHostPortEnv("ADDRESS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" || got.Port != nil {
			t.Fatalf("want empty host, nil port; got %+v", got)
		}
	})
}

func TestReadHostPortEnv_Valid(t *testing.T) {
	withEnv("ADDRESS", "example.com:8181", func() {
		got, err := ReadHostPortEnv("ADDRESS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "example.com" {
			t.Fatalf("host mismatch: %q", got.Host)
		}
		if got.Port == nil || *got.Port != 8181 {
			t.Fatalf("port mismatch: %+v", got.Port)
		}
	})
}

func TestReadHostPortEnv_EmptyHost(t *testing.T) {
	withEnv("ADDRESS", ":6060", func() {
		got, err := ReadHostPortEnv("ADDRESS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" || got.Port == nil || *got.Port != 6060 {
			t.Fatalf("want host='', port=6060; got %+v", got)
		}
	})
}

func TestReadHostPortEnv_InvalidIgnored(t *testing.T) {
	withEnv("ADDRESS", "not-an-addr", func() {
		got, err := ReadHostPortEnv("ADDRESS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" || got.Port != nil {
			t.Fatalf("invalid value must be ignored; got %+v", got)
		}
	})
}

func TestReadHostPortEnv_IPv6(t *testing.T) {
	withEnv("ADDRESS", "[2001:db8::1]:9091", func() {
		got, err := ReadHostPortEnv("ADDRESS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "2001:db8::1" {
			t.Fatalf("host mismatch: %q", got.Host)
		}
		if got.Port == nil || *got.Port != 9091 {
			t.Fatalf("port mismatch: %+v", got.Port)
		}
	})
}
