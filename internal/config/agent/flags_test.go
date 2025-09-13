package agentcfg

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
			t.Fatalf("want empty host, got %q", got.addressFlag.Host)
		}
		if got.addressFlag.Port != nil {
			t.Fatalf("want nil port, got %d", *got.addressFlag.Port)
		}
		if got.ReportIntervalSec != nil || got.PollIntervalSec != nil || got.RateLimit != nil {
			t.Fatalf("unexpected values: report=%v poll=%v limit=%v", got.ReportIntervalSec, got.PollIntervalSec, got.RateLimit)
		}
	})
}

func TestParseFlags_Address_Inline(t *testing.T) {
	withArgs([]string{"-a=1.2.3.4:9999"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.addressFlag.Host != "1.2.3.4" {
			t.Fatalf("host mismatch: %q", got.addressFlag.Host)
		}
		if got.addressFlag.Port == nil || *got.addressFlag.Port != 9999 {
			t.Fatalf("port mismatch: %v", got.addressFlag.Port)
		}
	})
}

func TestParseFlags_Address_Separate(t *testing.T) {
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

func TestParseFlags_Address_EmptyHost_AllIfaces(t *testing.T) {
	withArgs([]string{"-a", ":7777"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.addressFlag.Host != "" {
			t.Fatalf("want empty host, got %q", got.addressFlag.Host)
		}
		if got.addressFlag.Port == nil || *got.addressFlag.Port != 7777 {
			t.Fatalf("port mismatch: %v", got.addressFlag.Port)
		}
	})
}

func TestParseFlags_Address_IPv6_Bracketed(t *testing.T) {
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

func TestParseFlags_Address_Invalid_Error(t *testing.T) {
	withArgs([]string{"-a", "not-an-addr"}, func() {
		_, err := parseFlags()
		if err == nil {
			t.Fatalf("expected error for invalid -a")
		}
	})
}

func TestParseFlags_ReportPollLimit_OK(t *testing.T) {
	withArgs([]string{"-r", "15", "-p=3", "-l", "5"}, func() {
		got, err := parseFlags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ReportIntervalSec == nil || *got.ReportIntervalSec != 15 {
			t.Fatalf("report mismatch: %v", got.ReportIntervalSec)
		}
		if got.PollIntervalSec == nil || *got.PollIntervalSec != 3 {
			t.Fatalf("poll mismatch: %v", got.PollIntervalSec)
		}
		if got.RateLimit == nil || *got.RateLimit != 5 {
			t.Fatalf("limit mismatch: %v", got.RateLimit)
		}
	})
}

func TestParseFlags_Report_Invalid_Error(t *testing.T) {
	withArgs([]string{"-r", "-1"}, func() {
		_, err := parseFlags()
		if err == nil {
			t.Fatalf("expected error for invalid -r")
		}
	})
}

func TestParseFlags_Poll_Invalid_Error(t *testing.T) {
	withArgs([]string{"-p", "0"}, func() {
		_, err := parseFlags()
		if err == nil {
			t.Fatalf("expected error for invalid -p")
		}
	})
}

func TestParseFlags_UnknownFlag_Error(t *testing.T) {
	withArgs([]string{"-x"}, func() {
		_, err := parseFlags()
		if err == nil {
			t.Fatalf("expected parse error for unknown flag")
		}
	})
}

func TestFlagsValueMapper_Address_NonEmptyHost(t *testing.T) {
	var dst AgentFlags
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

func TestFlagsValueMapper_Address_EmptyHost_OnlyPortApplied(t *testing.T) {
	var dst AgentFlags
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

func TestFlagsValueMapper_ReportPollLimit(t *testing.T) {
	var dst AgentFlags

	rep := 12
	poll := 5
	lim := 7
	if err := flagsValueMapper(&dst, ReportSecondsFlagValue{Sec: &rep}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := flagsValueMapper(&dst, PollSecondsFlagValue{Sec: &poll}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := flagsValueMapper(&dst, RateLimitFlagValue{Rate: &lim}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dst.ReportIntervalSec == nil || *dst.ReportIntervalSec != 12 {
		t.Fatalf("report mismatch: %v", dst.ReportIntervalSec)
	}
	if dst.PollIntervalSec == nil || *dst.PollIntervalSec != 5 {
		t.Fatalf("poll mismatch: %v", dst.PollIntervalSec)
	}
	if dst.RateLimit == nil || *dst.RateLimit != 7 {
		t.Fatalf("limit mismatch: %v", dst.RateLimit)
	}
}

type unknownValue struct{ X int }

func TestFlagsValueMapper_UnknownType_Ignored(t *testing.T) {
	var dst AgentFlags
	if err := flagsValueMapper(&dst, unknownValue{X: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.addressFlag.Host != "" || dst.addressFlag.Port != nil || dst.ReportIntervalSec != nil || dst.PollIntervalSec != nil {
		t.Fatalf("dst must be unchanged on unknown type, got %+v", dst)
	}
}

func TestParseReportSecondsFlag_OK(t *testing.T) {
	v, err := ParseReportSecondsFlag("10", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Sec == nil || *v.Sec != 10 {
		t.Fatalf("mismatch: %+v", v)
	}
}

func TestParseReportSecondsFlag_NotPresent(t *testing.T) {
	v, err := ParseReportSecondsFlag("10", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Sec != nil {
		t.Fatalf("want nil when not present, got %+v", v)
	}
}

func TestParseReportSecondsFlag_Invalid(t *testing.T) {
	if _, err := ParseReportSecondsFlag("0", true); err == nil {
		t.Fatalf("expected error for 0")
	}
	if _, err := ParseReportSecondsFlag("-3", true); err == nil {
		t.Fatalf("expected error for negative")
	}
	if _, err := ParseReportSecondsFlag("NaN", true); err == nil {
		t.Fatalf("expected error for non-numeric")
	}
}

func TestParsePollSecondsFlag_OK(t *testing.T) {
	v, err := ParsePollSecondsFlag("2", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Sec == nil || *v.Sec != 2 {
		t.Fatalf("mismatch: %+v", v)
	}
}

func TestParsePollSecondsFlag_Invalid(t *testing.T) {
	if _, err := ParsePollSecondsFlag("0", true); err == nil {
		t.Fatalf("expected error for 0")
	}
}

func TestCloneMetrics_NilInput(t *testing.T) {

	if got := flagsValueMapper(nil, nil); got != nil {
		t.Errorf("expected nil, got %#v", got)
	}
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

func TestParseFlags_RateLimit_Invalid_Error(t *testing.T) {
	withArgs([]string{"-l", "0"}, func() {
		_, err := parseFlags()
		if err == nil {
			t.Fatalf("expected error for invalid -l")
		}
	})
}

func TestParseRateLimitFlag_OK(t *testing.T) {
	v, err := ParseRateLimitFlag("5", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Rate == nil || *v.Rate != 5 {
		t.Fatalf("mismatch: %+v", v)
	}
}

func TestParseRateLimitFlag_NotPresent(t *testing.T) {
	v, err := ParseRateLimitFlag("5", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Rate != nil {
		t.Fatalf("want nil when not present, got %+v", v)
	}
}

func TestParseRateLimitFlag_Invalid(t *testing.T) {
	if _, err := ParseRateLimitFlag("0", true); err == nil {
		t.Fatalf("expected error for 0")
	}
}
