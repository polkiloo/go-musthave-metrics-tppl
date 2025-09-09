package agentcfg

import (
	"os"
	"testing"
)

func withEnvMap(vars map[string]string, fn func()) {

	type oldEnt struct {
		val string
		ok  bool
	}
	old := map[string]oldEnt{}
	for k := range vars {
		v, ok := os.LookupEnv(k)
		old[k] = oldEnt{val: v, ok: ok}
	}
	for k, v := range vars {
		_ = os.Setenv(k, v)
	}
	fn()
	for k, e := range old {
		if e.ok {
			_ = os.Setenv(k, e.val)
		} else {
			_ = os.Unsetenv(k)
		}
	}
}

func TestGetEnvVars_NoVars(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "",
		EnvReportIntervalVarName: "",
		EnvPollIntervalVarName:   "",
	}, func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" {
			t.Fatalf("want empty host, got %q", got.Host)
		}
		if got.Port != nil || got.ReportIntervalSec != nil || got.PollIntervalSec != nil {
			t.Fatalf("want all nil pointers, got %+v", got)
		}
	})
}

func TestGetEnvVars_Address_Domain(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "example.com:8181",
		EnvReportIntervalVarName: "",
		EnvPollIntervalVarName:   "",
	}, func() {
		got, err := getEnvVars()
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

func TestGetEnvVars_Address_EmptyHost_AllIfaces(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName: ":6060",
	}, func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" {
			t.Fatalf("want empty host, got %q", got.Host)
		}
		if got.Port == nil || *got.Port != 6060 {
			t.Fatalf("port mismatch: %+v", got.Port)
		}
	})
}

func TestGetEnvVars_Address_IPv6Bracketed(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName: "[::1]:9090",
	}, func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "::1" {
			t.Fatalf("IPv6 host mismatch: %q", got.Host)
		}
		if got.Port == nil || *got.Port != 9090 {
			t.Fatalf("port mismatch: %+v", got.Port)
		}
	})
}

func TestGetEnvVars_Address_Invalid_Ignored(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName: "not-an-addr",
	}, func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "" || got.Port != nil {
			t.Fatalf("invalid ADDRESS must be ignored; got %+v", got)
		}
	})
}

func TestGetEnvVars_ReportAndPoll_Valid(t *testing.T) {
	withEnvMap(map[string]string{
		EnvReportIntervalVarName: "15",
		EnvPollIntervalVarName:   "3",
	}, func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ReportIntervalSec == nil || *got.ReportIntervalSec != 15 {
			t.Fatalf("report mismatch: %+v", got.ReportIntervalSec)
		}
		if got.PollIntervalSec == nil || *got.PollIntervalSec != 3 {
			t.Fatalf("poll mismatch: %+v", got.PollIntervalSec)
		}
	})
}

func TestGetEnvVars_Report_Invalid_Ignored(t *testing.T) {
	for _, v := range []string{"0", "-1", "NaN"} {
		withEnvMap(map[string]string{
			EnvReportIntervalVarName: v,
		}, func() {
			got, err := getEnvVars()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ReportIntervalSec != nil {
				t.Fatalf("invalid report must be ignored; got %v", *got.ReportIntervalSec)
			}
		})
	}
}

func TestGetEnvVars_Poll_Invalid_Ignored(t *testing.T) {
	for _, v := range []string{"0", "-2", "oops"} {
		withEnvMap(map[string]string{
			EnvPollIntervalVarName: v,
		}, func() {
			got, err := getEnvVars()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.PollIntervalSec != nil {
				t.Fatalf("invalid poll must be ignored; got %v", *got.PollIntervalSec)
			}
		})
	}
}

func TestGetEnvVars_Combined_AllSources(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "agent.local:8088",
		EnvReportIntervalVarName: "10",
		EnvPollIntervalVarName:   "2",
	}, func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Host != "agent.local" {
			t.Fatalf("host mismatch: %q", got.Host)
		}
		if got.Port == nil || *got.Port != 8088 {
			t.Fatalf("port mismatch: %+v", got.Port)
		}
		if got.ReportIntervalSec == nil || *got.ReportIntervalSec != 10 {
			t.Fatalf("report mismatch: %+v", got.ReportIntervalSec)
		}
		if got.PollIntervalSec == nil || *got.PollIntervalSec != 2 {
			t.Fatalf("poll mismatch: %+v", got.PollIntervalSec)
		}
	})
}

func TestGetEnvVars_Key(t *testing.T) {
	withEnvMap(map[string]string{EnvKeyVarName: "secret"}, func() {
		got, err := getEnvVars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SignKey != "secret" {
			t.Fatalf("key mismatch: %q", got.SignKey)
		}
	})
}
