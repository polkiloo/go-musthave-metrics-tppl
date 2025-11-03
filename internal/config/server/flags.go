package servercfg

import (
	"flag"
	"io"
	"os"
	"strconv"

	commoncfg "github.com/polkiloo/go-musthave-metrics-tppl/internal/config/common"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
)

type ServerFlags struct {
	addressFlag   commoncfg.AddressFlagValue
	storeInterval *int
	fileStorage   string
	restore       *bool
	SignKey       string
}

var (
	defaultAddress       = server.DefaultAppHost + ":" + strconv.Itoa(server.DefaultAppPort)
	defaultStoreInterval = server.DefaultStoreInterval
	defaultFileStorage   = server.DefaultFileStoragePath
	defaultRestore       = server.DefaultRestore
)

func parseFlags() (ServerFlags, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.String("a", defaultAddress, "HTTP endpoint, e.g., localhost:8080 or :8080")
	fs.Int("i", defaultStoreInterval, "store interval in seconds")
	fs.String("f", defaultFileStorage, "path to file for metrics storage")
	fs.Bool("r", defaultRestore, "restore metrics from file on start")
	fs.String("k", "", "key for hashing")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return ServerFlags{}, err
	}

	flags := ServerFlags{}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	if set["a"] {
		hp, err := commoncfg.ParseAddressFlag(fs.Lookup("a").Value.String(), true)
		if err != nil {
			return ServerFlags{}, err
		}
		flags.addressFlag = hp
	}
	if set["i"] {
		v, _ := strconv.Atoi(fs.Lookup("i").Value.String())
		flags.storeInterval = &v
	}
	if set["f"] {
		flags.fileStorage = fs.Lookup("f").Value.String()
	}
	if set["r"] {
		b, err := strconv.ParseBool(fs.Lookup("r").Value.String())
		if err != nil {
			return ServerFlags{}, err
		}
		flags.restore = &b
	}
	if set["k"] {
		flags.SignKey = fs.Lookup("k").Value.String()
	}

	return flags, nil
}
