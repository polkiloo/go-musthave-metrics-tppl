package agentcfg

import (
	"os"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	"go.uber.org/fx"
)

var defaultAppConfig = agent.AppConfig{
	Host:           agent.DefaultAppHost,
	Port:           agent.DefaultAppPort,
	PollInterval:   agent.DefaultAppPollInterval,
	ReportInterval: agent.DefaultAppReportInterval,
	LoopIterations: agent.DefaultLoopIterations,
	RateLimit:      agent.DefaultRateLimit,
	CryptoKeyPath:  agent.DefaultCryptoKeyPath,
}

func TestBuildAgentConfig_Default_WhenNoEnvNoFlags(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "",
		EnvReportIntervalVarName: "",
		EnvPollIntervalVarName:   "",
		EnvRateLimitVarName:      "",
	}, func() {
		withArgs(nil, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != defaultAppConfig.Host {
				t.Fatalf("want default host %q, got %q", defaultAppConfig.Host, got.Host)
			}
			if got.Port != defaultAppConfig.Port {
				t.Fatalf("want default port %d, got %d", defaultAppConfig.Port, got.Port)
			}
			if got.ReportInterval != defaultAppConfig.ReportInterval {
				t.Fatalf("want default report interval %v, got %v", defaultAppConfig.ReportInterval, got.ReportInterval)
			}
			if got.PollInterval != defaultAppConfig.PollInterval {
				t.Fatalf("want default poll interval %v, got %v", defaultAppConfig.PollInterval, got.PollInterval)
			}
			if got.RateLimit != defaultAppConfig.RateLimit {
				t.Fatalf("want default rate limit %d, got %d", defaultAppConfig.RateLimit, got.RateLimit)
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
		withArgs([]string{"-a", "flag-host:9090", "-r", "15", "-p", "3", "-l", "5"}, func() {
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
			if got.RateLimit != 5 {
				t.Fatalf("want ratelimit=5, got %d", got.RateLimit)
			}
		})
	})
}

func TestBuildAgentConfig_EnvOnly_EnvWins(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "env-host:8081",
		EnvReportIntervalVarName: "20",
		EnvPollIntervalVarName:   "4",
		EnvRateLimitVarName:      "7",
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
			if got.RateLimit != 7 {
				t.Fatalf("want ratelimit=7, got %d", got.RateLimit)
			}
		})
	})
}

func TestBuildAgentConfig_EnvBeatsFlags_ForEachField(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "env-host:8000",
		EnvReportIntervalVarName: "30",
		EnvPollIntervalVarName:   "6",
		EnvRateLimitVarName:      "9",
	}, func() {
		withArgs([]string{"-a", "flag-host:9000", "-r", "10", "-p", "2", "-l", "3"}, func() {
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
			if got.RateLimit != 9 {
				t.Fatalf("env must win for ratelimit=9, got %d", got.RateLimit)
			}
		})
	})
}

func TestBuildAgentConfig_MixedPerField_EnvPortFlagsHost(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:       ":7070",
		EnvRateLimitVarName:     "",
		EnvCryptoKeyPathVarName: "",
	}, func() {
		withArgs([]string{"-a", "flag-host:9999", "-l", "8", "-crypto-key", "flag-public.pem"}, func() {
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
			if got.RateLimit != 8 {
				t.Fatalf("want ratelimit from flags 8, got %d", got.RateLimit)
			}
			if got.CryptoKeyPath != "flag-public.pem" {
				t.Fatalf("want crypto key from flags 'flag-public.pem', got %q", got.CryptoKeyPath)
			}
		})
	})
}

func TestBuildAgentConfig_DefaultCryptoKeyEmpty(t *testing.T) {
	withEnvMap(map[string]string{
		EnvAddressVarName:        "",
		EnvReportIntervalVarName: "",
		EnvPollIntervalVarName:   "",
		EnvRateLimitVarName:      "",
		EnvCryptoKeyPathVarName:  "",
	}, func() {
		withArgs(nil, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.CryptoKeyPath != "" {
				t.Fatalf("default crypto key must be empty, got %q", got.CryptoKeyPath)
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
			if got.Host != defaultAppConfig.Host {
				t.Fatalf("empty host in flags must not override default; got %q", got.Host)
			}
			if got.Port != 9000 {
				t.Fatalf("want port=9000 from flags, got %d", got.Port)
			}
		})
	})
}

func TestBuildAgentConfig_ConfigFileAppliedBeforeEnvAndFlags(t *testing.T) {
	cfgFile := t.TempDir() + "/config.json"
	err := os.WriteFile(cfgFile, []byte(`{
                "address": "file-host:9000",
                "report_interval": "3s",
                "poll_interval": "2s",
                "key": "filekey",
                "rate_limit": 5,
                "crypto_key": "file-key.pem"
        }`), 0o600)
	if err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	withEnvMap(map[string]string{
		EnvAddressVarName:        "",
		EnvReportIntervalVarName: "",
		EnvPollIntervalVarName:   "",
		EnvRateLimitVarName:      "",
		EnvCryptoKeyPathVarName:  "",
		"CONFIG":                 cfgFile,
	}, func() {
		withArgs(nil, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != "file-host" || got.Port != 9000 {
				t.Fatalf("want address from file file-host:9000, got %s:%d", got.Host, got.Port)
			}
			if got.ReportInterval != 3*time.Second {
				t.Fatalf("want report interval 3s from file, got %v", got.ReportInterval)
			}
			if got.PollInterval != 2*time.Second {
				t.Fatalf("want poll interval 2s from file, got %v", got.PollInterval)
			}
			if got.SignKey != "filekey" {
				t.Fatalf("want key from file, got %q", got.SignKey)
			}
			if got.RateLimit != 5 {
				t.Fatalf("want rate limit 5 from file, got %d", got.RateLimit)
			}
			if got.CryptoKeyPath != "file-key.pem" {
				t.Fatalf("want crypto key from file, got %q", got.CryptoKeyPath)
			}
		})
	})
}

func TestBuildAgentConfig_EnvOverridesConfigFile(t *testing.T) {
	cfgFile := t.TempDir() + "/config.json"
	err := os.WriteFile(cfgFile, []byte(`{"address": "file-host:9000", "report_interval": "3s"}`), 0o600)
	if err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	withEnvMap(map[string]string{
		EnvAddressVarName:        "env-host:8085",
		EnvReportIntervalVarName: "8",
		"CONFIG":                 cfgFile,
	}, func() {
		withArgs([]string{"-a", "flag-host:9999"}, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != "env-host" || got.Port != 8085 {
				t.Fatalf("env should override config file, got %s:%d", got.Host, got.Port)
			}
			if got.ReportInterval != 8*time.Second {
				t.Fatalf("env report interval should override file, got %v", got.ReportInterval)
			}
		})
	})
}

func TestAgentConfigModule_ValidateApp(t *testing.T) {
	if err := fx.ValidateApp(Module); err != nil {
		t.Fatalf("fx.ValidateApp(Module) failed: %v", err)
	}
}

func TestBuildAgentConfig_KeyPriority(t *testing.T) {
	withEnvMap(map[string]string{EnvKeyVarName: "envkey"}, func() {
		withArgs([]string{"-k", "flagkey"}, func() {
			got, err := buildAgentConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.SignKey != "envkey" {
				t.Fatalf("env key must win: got %q", got.SignKey)
			}
		})
	})
}
