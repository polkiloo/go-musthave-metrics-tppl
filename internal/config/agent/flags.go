package agentcfg

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/agent"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
)

var (
	ErrUnknownArgs = fmt.Errorf("unknown args")
)

type AgentFlags struct {
	addressFlag       commoncfg.AddressFlagValue
	ReportIntervalSec *int // nil если не задано
	PollIntervalSec   *int // nil если не задано
}

var (
	defaultAddress = agent.DefaultAppHost + ":" + strconv.Itoa(agent.DefaultAppPort)
)

type ReportSecondsFlagValue struct{ Sec *int }
type PollSecondsFlagValue struct{ Sec *int }

func ParseReportSecondsFlag(value string, present bool) (ReportSecondsFlagValue, error) {
	if !present {
		return ReportSecondsFlagValue{}, nil
	}
	sec, err := strconv.Atoi(value)
	if err != nil || sec <= 0 {
		return ReportSecondsFlagValue{}, fmt.Errorf("invalid -r (reportInterval): %q", value)
	}
	return ReportSecondsFlagValue{Sec: &sec}, nil
}

func ParsePollSecondsFlag(value string, present bool) (PollSecondsFlagValue, error) {
	if !present {
		return PollSecondsFlagValue{}, nil
	}
	sec, err := strconv.Atoi(value)
	if err != nil || sec <= 0 {
		return PollSecondsFlagValue{}, fmt.Errorf("invalid -p (pollInterval): %q", value)
	}
	return PollSecondsFlagValue{Sec: &sec}, nil
}

func flagsValueMapper(dst *AgentFlags, v commoncfg.FlagValue) error {
	switch t := v.(type) {
	case nil:
		return nil
	case commoncfg.AddressFlagValue:
		if t.Host != "" {
			dst.addressFlag.Host = t.Host
		}
		if t.Port != nil {
			dst.addressFlag.Port = t.Port
		}
		return nil
	case ReportSecondsFlagValue:
		if t.Sec != nil {
			dst.ReportIntervalSec = t.Sec
		}
		return nil
	case PollSecondsFlagValue:
		if t.Sec != nil {
			dst.PollIntervalSec = t.Sec
		}
		return nil
	default:
		return nil
	}
}

func parseFlags() (AgentFlags, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [-a host:port] [-r seconds] [-p seconds]\n", os.Args[0])
		fs.PrintDefaults()
	}

	fs.String("a", defaultAddress, "HTTP endpoint, e.g., localhost:8080 or :8080")
	fs.String("r", "", "reportInterval in seconds (default 10)")
	fs.String("p", "", "pollInterv1al in seconds (default 2)")

	return commoncfg.
		NewDispatcher[AgentFlags](fs, flagsValueMapper).
		Handle("a", commoncfg.Lift(commoncfg.ParseAddressFlag)).
		Handle("r", commoncfg.Lift(ParseReportSecondsFlag)).
		Handle("p", commoncfg.Lift(ParsePollSecondsFlag)).
		Parse(os.Args[1:])
}
