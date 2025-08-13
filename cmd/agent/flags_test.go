package main

import (
	"flag"
	"net"
	"os"
	"testing"
)

func TestParseFlags_Defaults_NoFlags(t *testing.T) {
	resetFlags()
	defer setArgs()()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Host != "" {
		t.Fatalf("Host must be empty when address is default, got %q", got.Host)
	}
	if got.Port != nil {
		t.Fatalf("Port must be nil when address is default, got %v", *got.Port)
	}
	if got.ReportIntervalSec != nil {
		t.Fatalf("ReportIntervalSec must be nil by default, got %v", *got.ReportIntervalSec)
	}
	if got.PollIntervalSec != nil {
		t.Fatalf("PollIntervalSec must be nil by default, got %v", *got.PollIntervalSec)
	}
}

func TestParseFlags_UnknownArgs(t *testing.T) {
	resetFlags()
	defer setArgs("positional")()

	_, err := parseFlags()
	if err == nil {
		t.Fatalf("expected error for unknown args, got nil")
	}
	if err != ErrUnknownArgs {
		t.Fatalf("expected ErrUnknownArgs, got %v", err)
	}
}

func TestParseFlags_ValidAddressOnly(t *testing.T) {
	resetFlags()
	defer setArgs("-a=myhost:4321")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "myhost" {
		t.Fatalf("Host mismatch: want %q, got %q", "myhost", got.Host)
	}
	if got.Port == nil || *got.Port != 4321 {
		if got.Port == nil {
			t.Fatalf("Port is nil, want 4321")
		}
		t.Fatalf("Port mismatch: want 4321, got %d", *got.Port)
	}
	if got.ReportIntervalSec != nil || got.PollIntervalSec != nil {
		t.Fatalf("intervals must be nil when not provided; got r=%v p=%v", got.ReportIntervalSec, got.PollIntervalSec)
	}
}

func TestParseFlags_InvalidAddress_NoPort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=justhost")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("did not expect error, got %v (function prints but does not return error on invalid -a)", err)
	}
	if got.Host != "" {
		t.Fatalf("Host must be empty on invalid -a, got %q", got.Host)
	}
	if got.Port != nil {
		t.Fatalf("Port must be nil on invalid -a, got %v", *got.Port)
	}
}

func TestParseFlags_InvalidAddress_NonNumericPort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=host:notaport")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	if got.Host != "host" {
		t.Fatalf("Host mismatch: want host, got %q", got.Host)
	}
	if got.Port != nil {
		t.Fatalf("Port must be nil for non-numeric port, got %v", *got.Port)
	}
}

func TestParseFlags_InvalidAddress_ZeroPort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=host:0")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	if got.Host != "host" {
		t.Fatalf("Host mismatch: want host, got %q", got.Host)
	}
	if got.Port != nil {
		t.Fatalf("Port must be nil for zero port, got %v", *got.Port)
	}
}

func TestParseFlags_ReportInterval_Valid(t *testing.T) {
	resetFlags()
	defer setArgs("-r=15")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ReportIntervalSec == nil || *got.ReportIntervalSec != 15 {
		if got.ReportIntervalSec == nil {
			t.Fatalf("ReportIntervalSec is nil, want 15")
		}
		t.Fatalf("ReportIntervalSec mismatch: want 15, got %d", *got.ReportIntervalSec)
	}
	if got.Host != "" || got.Port != nil {
		t.Fatalf("address should remain default; got host=%q port=%v", got.Host, got.Port)
	}
}

func TestParseFlags_ReportInterval_Invalid_NotNumber(t *testing.T) {
	resetFlags()
	defer setArgs("-r=notanumber")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	if got.ReportIntervalSec == nil {
		t.Fatalf("ReportIntervalSec must be set (even if invalid), got nil")
	}
	if *got.ReportIntervalSec != 0 {
		t.Fatalf("ReportIntervalSec must be 0 when invalid input, got %d", *got.ReportIntervalSec)
	}
}

func TestParseFlags_ReportInterval_Invalid_Zero(t *testing.T) {
	resetFlags()
	defer setArgs("-r=0")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	if got.ReportIntervalSec == nil || *got.ReportIntervalSec != 0 {
		if got.ReportIntervalSec == nil {
			t.Fatalf("ReportIntervalSec must be set")
		}
		t.Fatalf("ReportIntervalSec must be 0 for zero input, got %d", *got.ReportIntervalSec)
	}
}

func TestParseFlags_PollInterval_Valid(t *testing.T) {
	resetFlags()
	defer setArgs("-p=7")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PollIntervalSec == nil || *got.PollIntervalSec != 7 {
		if got.PollIntervalSec == nil {
			t.Fatalf("PollIntervalSec is nil, want 7")
		}
		t.Fatalf("PollIntervalSec mismatch: want 7, got %d", *got.PollIntervalSec)
	}
}

func TestParseFlags_PollInterval_Invalid_Negative(t *testing.T) {
	resetFlags()
	defer setArgs("-p=-3")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	if got.PollIntervalSec == nil || *got.PollIntervalSec != -3 {
		if got.PollIntervalSec == nil {
			t.Fatalf("PollIntervalSec must be set")
		}
		t.Fatalf("PollIntervalSec must be -3 for input -3, got %d", *got.PollIntervalSec)
	}
}

func TestParseFlags_AllFlagsTogether(t *testing.T) {
	resetFlags()
	defer setArgs("-a=0.0.0.0:9090", "-r=11", "-p=4")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Host != "0.0.0.0" {
		t.Fatalf("Host mismatch: want 0.0.0.0, got %q", got.Host)
	}
	if got.Port == nil || *got.Port != 9090 {
		if got.Port == nil {
			t.Fatalf("Port is nil, want 9090")
		}
		t.Fatalf("Port mismatch: want 9090, got %d", *got.Port)
	}
	if got.ReportIntervalSec == nil || *got.ReportIntervalSec != 11 {
		if got.ReportIntervalSec == nil {
			t.Fatalf("ReportIntervalSec is nil, want 11")
		}
		t.Fatalf("ReportIntervalSec mismatch: want 11, got %d", *got.ReportIntervalSec)
	}
	if got.PollIntervalSec == nil || *got.PollIntervalSec != 4 {
		if got.PollIntervalSec == nil {
			t.Fatalf("PollIntervalSec is nil, want 4")
		}
		t.Fatalf("PollIntervalSec mismatch: want 4, got %d", *got.PollIntervalSec)
	}
}

func TestDefaultAddress_IsValidHostPort(t *testing.T) {
	_, _, err := net.SplitHostPort(defaultAddress)
	if err != nil {
		t.Fatalf("defaultAddress %q must be a valid host:port, err=%v", defaultAddress, err)
	}
}

func TestErrUnknownArgsReturned(t *testing.T) {
	resetFlags()
	defer setArgs("x")()

	_, err := parseFlags()
	if err == nil || err != ErrUnknownArgs {
		t.Fatalf("expected ErrUnknownArgs, got %v", err)
	}
}

func TestParseFlags_Address_MinPort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=host:1")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "host" {
		t.Fatalf("Host mismatch: want host, got %q", got.Host)
	}
	if got.Port == nil || *got.Port != 1 {
		t.Fatalf("Port mismatch: want 1, got %v", got.Port)
	}
}

func TestParseFlags_Address_LargePort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=host:65535")()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Port == nil || *got.Port != 65535 {
		t.Fatalf("Port mismatch: want 65535, got %v", got.Port)
	}
}

func TestParseFlags_Address_EqualsDefault(t *testing.T) {
	resetFlags()
	defer setArgs("-a=" + defaultAddress)()

	got, err := parseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Host != "" || got.Port != nil {
		t.Fatalf("Host/Port must not be set when address equals default; got host=%q port=%v", got.Host, got.Port)
	}
}

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func setArgs(args ...string) func() {
	orig := os.Args
	os.Args = append([]string{orig[0]}, args...)
	return func() { os.Args = orig }
}

func intPtr(v int) *int { return &v }
