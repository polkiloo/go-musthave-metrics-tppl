package servercfg

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
)

type ServerFlags struct {
	addressFlag commoncfg.AddressFlagValue
}

var (
	defaultAddress = server.DefaultAppHost + ":" + strconv.Itoa(server.DefaultAppPort)
)

func parseFlags() (ServerFlags, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [-a host:port]\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.String("a", defaultAddress, "HTTP endpoint, e.g., localhost:8080 or :8080")

	res, err := commoncfg.
		NewDispatcher[ServerFlags](fs, flagsValueMapper).
		Handle("a", commoncfg.Lift(commoncfg.ParseAddressFlag)).
		Parse(os.Args[1:])

	if err != nil {
		return ServerFlags{}, err
	}

	return res, nil
}

func flagsValueMapper(dst *ServerFlags, v commoncfg.FlagValue) error {
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
	default:
		return nil
	}
}
