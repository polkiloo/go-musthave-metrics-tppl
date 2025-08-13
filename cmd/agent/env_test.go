package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead_AllEnvVarsSet(t *testing.T) {
	os.Setenv("ADDRESS", "localhost:8080")
	os.Setenv("REPORT_INTERVAL", "10")
	os.Setenv("POLL_INTERVAL", "5")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	env := getEnvVars()
	assert.Equal(t, "localhost", env.Host)
	assert.NotNil(t, env.Port)
	assert.Equal(t, 8080, *env.Port)
	assert.NotNil(t, env.ReportIntervalSec)
	assert.Equal(t, 10, *env.ReportIntervalSec)
	assert.NotNil(t, env.PollIntervalSec)
	assert.Equal(t, 5, *env.PollIntervalSec)
}

func TestRead_OnlyAddress(t *testing.T) {
	os.Setenv("ADDRESS", "host:1234")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")
	defer os.Unsetenv("ADDRESS")

	env := getEnvVars()
	assert.Equal(t, "host", env.Host)
	assert.NotNil(t, env.Port)
	assert.Equal(t, 1234, *env.Port)
	assert.Nil(t, env.ReportIntervalSec)
	assert.Nil(t, env.PollIntervalSec)
}

func TestRead_OnlyIntervals(t *testing.T) {
	os.Unsetenv("ADDRESS")
	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("POLL_INTERVAL", "7")
	defer func() {
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	env := getEnvVars()
	assert.Equal(t, "", env.Host)
	assert.Nil(t, env.Port)
	assert.NotNil(t, env.ReportIntervalSec)
	assert.Equal(t, 15, *env.ReportIntervalSec)
	assert.NotNil(t, env.PollIntervalSec)
	assert.Equal(t, 7, *env.PollIntervalSec)
}

func TestRead_NoEnvVars(t *testing.T) {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")

	env := getEnvVars()
	assert.Equal(t, "", env.Host)
	assert.Nil(t, env.Port)
	assert.Nil(t, env.ReportIntervalSec)
	assert.Nil(t, env.PollIntervalSec)
}

func TestRead_InvalidEnvVars(t *testing.T) {
	os.Setenv("ADDRESS", "badaddress")
	os.Setenv("REPORT_INTERVAL", "notanumber")
	os.Setenv("POLL_INTERVAL", "-1")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	env := getEnvVars()
	assert.Equal(t, "", env.Host)
	assert.Nil(t, env.Port)
	assert.Nil(t, env.ReportIntervalSec)
	assert.Nil(t, env.PollIntervalSec)
}

func TestRead_ZeroAndNegativeIntervals(t *testing.T) {
	os.Setenv("REPORT_INTERVAL", "0")
	os.Setenv("POLL_INTERVAL", "-5")
	defer func() {
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	env := getEnvVars()
	assert.Nil(t, env.ReportIntervalSec)
	assert.Nil(t, env.PollIntervalSec)
}

func Test_splitHostPort_Valid(t *testing.T) {
	host, port, ok := splitHostPort("myhost:12345")
	assert.True(t, ok)
	assert.Equal(t, "myhost", host)
	assert.Equal(t, 12345, port)
}

func Test_splitHostPort_InvalidCases(t *testing.T) {
	tests := []struct {
		addr     string
		wantHost string
		wantPort int
		wantOK   bool
	}{
		{"host:", "", 0, false},
		{"host:notaport", "", 0, false},
		{"host:0", "", 0, false},
		{":1234", "", 0, false},
		{"host:-1", "", 0, false},
		{"", "", 0, false},
		{"host", "", 0, false},
	}

	for _, tt := range tests {
		h, p, ok := splitHostPort(tt.addr)
		assert.Equal(t, tt.wantHost, h)
		assert.Equal(t, tt.wantPort, p)
		assert.Equal(t, tt.wantOK, ok)
	}
}
