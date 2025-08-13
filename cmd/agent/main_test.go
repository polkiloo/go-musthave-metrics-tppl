package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"
	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestBuildAppConfig_Defaults(t *testing.T) {
	env := EnvVars{}
	fl := FlagsArg{}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, agent.DefaultAppConfig.Host, got.Host)
	assert.Equal(t, agent.DefaultAppConfig.Port, got.Port)
	assert.Equal(t, agent.DefaultAppConfig.PollInterval, got.PollInterval)
	assert.Equal(t, agent.DefaultAppConfig.ReportInterval, got.ReportInterval)
}

func TestBuildAppConfig_Host_EnvOverridesFlagAndDefault(t *testing.T) {
	env := EnvVars{Host: "env-host"}
	fl := FlagsArg{Host: "flag-host"}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, "env-host", got.Host)
}

func TestBuildAppConfig_Host_FlagOverridesDefault_WhenNoEnv(t *testing.T) {
	env := EnvVars{}
	fl := FlagsArg{Host: "flag-host"}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, "flag-host", got.Host)
}

func TestBuildAppConfig_Port_EnvOverridesFlagAndDefault(t *testing.T) {
	env := EnvVars{Port: i(9001)}
	fl := FlagsArg{Port: i(7777)}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, 9001, got.Port)
}

func TestBuildAppConfig_Port_FlagOverridesDefault_WhenNoEnv(t *testing.T) {
	env := EnvVars{}
	fl := FlagsArg{Port: i(7777)}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, 7777, got.Port)
}

func TestBuildAppConfig_PollInterval_EnvOverridesFlagAndDefault(t *testing.T) {
	env := EnvVars{PollIntervalSec: i(5)}
	fl := FlagsArg{PollIntervalSec: i(2)}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, got.PollInterval)
}

func TestBuildAppConfig_PollInterval_FlagOverridesDefault_WhenNoEnv(t *testing.T) {
	env := EnvVars{}
	fl := FlagsArg{PollIntervalSec: i(3)}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, 3*time.Second, got.PollInterval)
}

func TestBuildAppConfig_ReportInterval_EnvOverridesFlagAndDefault(t *testing.T) {
	env := EnvVars{ReportIntervalSec: i(11)}
	fl := FlagsArg{ReportIntervalSec: i(7)}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, 11*time.Second, got.ReportInterval)
}

func TestBuildAppConfig_ReportInterval_FlagOverridesDefault_WhenNoEnv(t *testing.T) {
	env := EnvVars{}
	fl := FlagsArg{ReportIntervalSec: i(13)}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)
	assert.Equal(t, 13*time.Second, got.ReportInterval)
}

func TestBuildAppConfig_MixedSources(t *testing.T) {
	env := EnvVars{
		Host:              "env-host",
		ReportIntervalSec: i(20),
	}
	fl := FlagsArg{
		Port:            i(5050),
		PollIntervalSec: i(4),
	}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)

	assert.Equal(t, "env-host", got.Host)
	assert.Equal(t, 5050, got.Port)
	assert.Equal(t, 4*time.Second, got.PollInterval)
	assert.Equal(t, 20*time.Second, got.ReportInterval)
}

func TestBuildAppConfig_PartialPointersNil(t *testing.T) {
	env := EnvVars{
		PollIntervalSec:   nil,
		ReportIntervalSec: i(9),
	}
	fl := FlagsArg{
		PollIntervalSec: i(6),
	}

	got, err := buildAppConfig(env, fl)
	assert.NoError(t, err)

	assert.Equal(t, agent.DefaultAppConfig.Host, got.Host)
	assert.Equal(t, agent.DefaultAppConfig.Port, got.Port)
	assert.Equal(t, 6*time.Second, got.PollInterval)
	assert.Equal(t, 9*time.Second, got.ReportInterval)
}
func TestBuildApp_StartStop_CoversMainPath(t *testing.T) {
	cfg := agent.AppConfig{
		Host:           "127.0.0.1",
		Port:           8080,
		PollInterval:   2 * time.Millisecond,
		ReportInterval: 3 * time.Millisecond,
	}
	tc := &mockCollector{}
	ts := &mockSender{}

	app := buildApp(
		cfg,
		fx.Decorate(func(agent.CollectorInterface) agent.CollectorInterface { return tc }),
		fx.Decorate(func(agent.SenderInterface) agent.SenderInterface { return ts }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	require.NoError(t, app.Start(ctx))
	time.Sleep(20 * time.Millisecond)
	require.NoError(t, app.Stop(ctx))

	require.Greater(t, atomic.LoadInt32(&tc.collects), int32(0))
	require.Greater(t, atomic.LoadInt32(&ts.sends), int32(0))
}

type mockCollector struct{ collects int32 }

func (m *mockCollector) Collect() {
	atomic.AddInt32(&m.collects, 1)
}
func (m *mockCollector) Snapshot() (map[string]models.Gauge, map[string]models.Counter) {
	return map[string]models.Gauge{"Alloc": 1.23}, map[string]models.Counter{"PollCount": 2}
}

type mockSender struct{ sends int32 }

func (m *mockSender) Send(g map[string]models.Gauge, c map[string]models.Counter) {
	atomic.AddInt32(&m.sends, 1)
}

func i(v int) *int { return &v }
