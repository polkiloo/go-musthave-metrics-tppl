package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func setArgs(args ...string) func() {
	orig := os.Args
	os.Args = append([]string{orig[0]}, args...)
	return func() { os.Args = orig }
}

func TestParseFlags_Defaults(t *testing.T) {
	resetFlags()
	defer setArgs()()
	args, err := parseFlags()
	assert.NoError(t, err)
	assert.Equal(t, defaultHost, args.Host)
	assert.Equal(t, defaultPort, args.Port)
	assert.Equal(t, defaultReportInterval, args.ReportInterval)
	assert.Equal(t, defaultPollInterval, args.PollInterval)
}

func TestParseFlags_AllFlags(t *testing.T) {
	resetFlags()
	defer setArgs("-a=myhost:4321", "-r=20", "-p=7")()
	args, err := parseFlags()
	assert.NoError(t, err)
	assert.Equal(t, "myhost", args.Host)
	assert.Equal(t, 4321, args.Port)
	assert.Equal(t, 20*time.Second, args.ReportInterval)
	assert.Equal(t, 7*time.Second, args.PollInterval)
}

func TestParseFlags_OnlyIntervals(t *testing.T) {
	resetFlags()
	defer setArgs("-r=22", "-p=12")()
	args, err := parseFlags()
	assert.NoError(t, err)
	assert.Equal(t, defaultHost, args.Host)
	assert.Equal(t, defaultPort, args.Port)
	assert.Equal(t, 22*time.Second, args.ReportInterval)
	assert.Equal(t, 12*time.Second, args.PollInterval)
}

func TestParseFlags_InvalidAddress_NoPort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=justhost")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrParseAddressFlags)
}

func TestParseFlags_InvalidAddress_EmptyPort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=host:")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrParseAddressFlags)
}

func TestParseFlags_InvalidAddress_NonNumericPort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=host:notaport")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrParseAddressFlags)
}

func TestParseFlags_InvalidAddress_ZeroPort(t *testing.T) {
	resetFlags()
	defer setArgs("-a=host:0")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrParseAddressFlags)
}

func TestParseFlags_InvalidReportInterval_NotNumber(t *testing.T) {
	resetFlags()
	defer setArgs("-r=notanumber")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrParseReportIntervalFlags)
}

func TestParseFlags_InvalidReportInterval_Zero(t *testing.T) {
	resetFlags()
	defer setArgs("-r=0")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrParseReportIntervalFlags)
}

func TestParseFlags_InvalidPollInterval_Negative(t *testing.T) {
	resetFlags()
	defer setArgs("-p=-3")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrParsePollIntervalFlags)
}

func TestParseFlags_InvalidPollInterval_NotNumber(t *testing.T) {
	resetFlags()
	defer setArgs("-p=oops")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrParsePollIntervalFlags)
}

func TestParseFlags_UnknownArgs(t *testing.T) {
	resetFlags()
	defer setArgs("extraneous")()
	_, err := parseFlags()
	assert.ErrorIs(t, err, ErrUnknownAddress)
}
