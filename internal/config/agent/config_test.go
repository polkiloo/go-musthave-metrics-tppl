package agentcfg

import (
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"go.uber.org/fx"
)

func TestBuildAgentConfig_Default_WhenNoEnvNoFlags(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "",
		EnvReportIntervalVarName: "",
		EnvPollIntervalVarName:   "",
	}, func() {
		withArgs(nil, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != agent.DefaultAppConfig.Host {
				t.Fatalf("want default host %q, got %q", agent.DefaultAppConfig.Host, got.Host)
			}
			if got.Port != agent.DefaultAppConfig.Port {
				t.Fatalf("want default port %d, got %d", agent.DefaultAppConfig.Port, got.Port)
			}
			if got.ReportInterval != agent.DefaultAppConfig.ReportInterval {
				t.Fatalf("want default report interval %v, got %v", agent.DefaultAppConfig.ReportInterval, got.ReportInterval)
			}
			if got.PollInterval != agent.DefaultAppConfig.PollInterval {
				t.Fatalf("want default poll interval %v, got %v", agent.DefaultAppConfig.PollInterval, got.PollInterval)
			}
		})
	})
}

func TestBuildAgentConfig_FlagsOnly_UsedWhenNoEnv(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "",
		EnvReportIntervalVarName: "",
		EnvPollIntervalVarName:   "",
	}, func() {
		withArgs([]string{"-a", "flag-host:9090", "-r", "15", "-p", "3"}, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != "flag-host" {
				t.Fatalf("want host from flags %q, got %q", "flag-host", got.Host)
			}
			if got.Port != 9090 {
				t.Fatalf("want port from flags 9090, got %d", got.Port)
			}
			if got.ReportInterval != 15*time.Second {
				t.Fatalf("want report=15s, got %v", got.ReportInterval)
			}
			if got.PollInterval != 3*time.Second {
				t.Fatalf("want poll=3s, got %v", got.PollInterval)
			}
		})
	})
}

func TestBuildAgentConfig_EnvOnly_EnvWins(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "env-host:8081",
		EnvReportIntervalVarName: "20",
		EnvPollIntervalVarName:   "4",
	}, func() {
		withArgs(nil, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != "env-host" || got.Port != 8081 {
				t.Fatalf("want env host/port env-host:8081, got %s:%d", got.Host, got.Port)
			}
			if got.ReportInterval != 20*time.Second {
				t.Fatalf("want report=20s, got %v", got.ReportInterval)
			}
			if got.PollInterval != 4*time.Second {
				t.Fatalf("want poll=4s, got %v", got.PollInterval)
			}
		})
	})
}

func TestBuildAgentConfig_EnvBeatsFlags_ForEachField(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "env-host:8000",
		EnvReportIntervalVarName: "30",
		EnvPollIntervalVarName:   "6",
	}, func() {
		withArgs([]string{"-a", "flag-host:9000", "-r", "10", "-p", "2"}, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != "env-host" || got.Port != 8000 {
				t.Fatalf("env must win for host/port, got %s:%d", got.Host, got.Port)
			}
			if got.ReportInterval != 30*time.Second {
				t.Fatalf("env must win for report=30s, got %v", got.ReportInterval)
			}
			if got.PollInterval != 6*time.Second {
				t.Fatalf("env must win for poll=6s, got %v", got.PollInterval)
			}
		})
	})
}

func TestBuildAgentConfig_MixedPerField_EnvPortFlagsHost(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName: ":7070",
	}, func() {
		withArgs([]string{"-a", "flag-host:9999"}, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != "flag-host" {
				t.Fatalf("want host from flags 'flag-host', got %q", got.Host)
			}
			if got.Port != 7070 {
				t.Fatalf("want port from env 7070, got %d", got.Port)
			}
		})
	})
}

func TestBuildAgentConfig_FlagEmptyHost_KeepDefaultHost_PortFromFlag(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "",
		EnvReportIntervalVarName: "",
		EnvPollIntervalVarName:   "",
	}, func() {
		withArgs([]string{"-a", ":9000"}, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != agent.DefaultAppConfig.Host {
				t.Fatalf("empty host in flags must not override default; got %q", got.Host)
			}
			if got.Port != 9000 {
				t.Fatalf("want port=9000 from flags, got %d", got.Port)
			}
		})
	})
}

/* ---------- fx Module wiring ---------- */

func TestAgentConfigModule_ValidateApp(t *testing.T) {
	if err := fx.ValidateApp(Module); err != nil {
		t.Fatalf("fx.ValidateApp(Module) failed: %v", err)
	}
}
