package commoncfg

import (
	"errors"
	"testing"
)

func TestParseAddressFlag_NotPresent(t *testing.T) {
	got, err := ParseAddressFlag("ignored:1234", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "" {
		t.Fatalf("want empty host when not present, got %q", got.Host)
	}
	if got.Port != nil {
		t.Fatalf("want nil port when not present, got %v", *got.Port)
	}
}

func TestParseAddressFlag_ValidHostPort(t *testing.T) {
	got, err := ParseAddressFlag("localhost:8080", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "localhost" {
		t.Fatalf("host mismatch: %q", got.Host)
	}
	if got.Port == nil || *got.Port != 8080 {
		t.Fatalf("port mismatch: %+v", got.Port)
	}
}

func TestParseAddressFlag_ValidEmptyHost_AllIfaces(t *testing.T) {
	got, err := ParseAddressFlag(":9090", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "" {
		t.Fatalf("want empty host, got %q", got.Host)
	}
	if got.Port == nil || *got.Port != 9090 {
		t.Fatalf("port mismatch: %+v", got.Port)
	}
}

func TestParseAddressFlag_ValidIPv6Bracketed(t *testing.T) {
	got, err := ParseAddressFlag("[::1]:7070", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "::1" {
		t.Fatalf("IPv6 host mismatch: %q", got.Host)
	}
	if got.Port == nil || *got.Port != 7070 {
		t.Fatalf("port mismatch: %+v", got.Port)
	}
}

func TestParseAddressFlag_Invalid_NoPortPart(t *testing.T) {
	// отсутствует двоеточие -> SplitHostPort вернёт ошибку
	_, err := ParseAddressFlag("localhost", true)
	if !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("want ErrInvalidAddress, got %v", err)
	}
}

func TestParseAddressFlag_Invalid_EmptyPort(t *testing.T) {
	// "localhost:" -> host есть, portStr пуст -> ErrInvalidPort
	_, err := ParseAddressFlag("localhost:", true)
	if !errors.Is(err, ErrInvalidPort) {
		t.Fatalf("want ErrInvalidPort, got %v", err)
	}
}

func TestParseAddressFlag_Invalid_NonNumericPort(t *testing.T) {
	_, err := ParseAddressFlag("127.0.0.1:http", true)
	if !errors.Is(err, ErrInvalidPort) {
		t.Fatalf("want ErrInvalidPort, got %v", err)
	}
}

func TestParseAddressFlag_Invalid_ZeroPort(t *testing.T) {
	_, err := ParseAddressFlag("127.0.0.1:0", true)
	if !errors.Is(err, ErrInvalidPort) {
		t.Fatalf("want ErrInvalidPort, got %v", err)
	}
}

func TestParseAddressFlag_Invalid_NegativePort(t *testing.T) {
	_, err := ParseAddressFlag("127.0.0.1:-1", true)
	if !errors.Is(err, ErrInvalidPort) {
		t.Fatalf("want ErrInvalidPort, got %v", err)
	}
}
