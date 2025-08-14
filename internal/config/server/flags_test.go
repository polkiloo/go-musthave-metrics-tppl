package servercfg

import (
	"os"
	"testing"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
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

func TestParseFlags_PositionalArg_Error(t *testing.T) {
	withArgs([]string{"-a", "localhost:1234", "positional"}, func() {
		_, err := parseFlags()
		if err == nil {
			t.Fatalf("expected ErrUnknownArgs")
		}
	})
}

func TestFlagsValueMapper_Address_NonEmptyHost(t *testing.T) {
	var dst ServerFlags
	p := 9090
	v := commoncfg.AddressFlagValue{Host: "example.com", Port: &p}
	if err := flagsValueMapper(&dst, v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.addressFlag.Host != "example.com" {
		t.Fatalf("host not applied: %q", dst.addressFlag.Host)
	}
	if dst.addressFlag.Port == nil || *dst.addressFlag.Port != 9090 {
		t.Fatalf("port not applied: %v", dst.addressFlag.Port)
	}
}

func TestFlagsValueMapper_Address_EmptyHost(t *testing.T) {
	var dst ServerFlags
	p := 6060
	v := commoncfg.AddressFlagValue{Host: "", Port: &p}
	if err := flagsValueMapper(&dst, v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.addressFlag.Host != "" {
		t.Fatalf("empty host must not override, got %q", dst.addressFlag.Host)
	}
	if dst.addressFlag.Port == nil || *dst.addressFlag.Port != 6060 {
		t.Fatalf("port not applied: %v", dst.addressFlag.Port)
	}
}

func TestFlagsValueMapper_NilValue_NoChange(t *testing.T) {
	var dst ServerFlags
	if err := flagsValueMapper(&dst, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.addressFlag.Host != "" || dst.addressFlag.Port != nil {
		t.Fatalf("dst must be unchanged, got %+v", dst)
	}
}

type unknownValue struct{ X int }

func TestFlagsValueMapper_UnknownType_Ignored(t *testing.T) {
	var dst ServerFlags
	if err := flagsValueMapper(&dst, unknownValue{X: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.addressFlag.Host != "" || dst.addressFlag.Port != nil {
		t.Fatalf("dst must be unchanged on unknown type, got %+v", dst)
	}
}
