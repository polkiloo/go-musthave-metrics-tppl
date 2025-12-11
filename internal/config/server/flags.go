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
	auditFile     string
	auditURL      string
	CryptoKeyPath string
	ConfigPath    string
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
	fs.String("audit-file", "", "path to audit log file")
	fs.String("audit-url", "", "remote URL for audit events")
	fs.String("crypto-key", "", "path to private key for decryption")
	fs.String("c", "", "path to configuration file")
	fs.String("config", "", "path to configuration file")

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

	if set["audit-file"] {
		flags.auditFile = fs.Lookup("audit-file").Value.String()
	}
	if set["audit-url"] {
		flags.auditURL = fs.Lookup("audit-url").Value.String()
	}

	if set["crypto-key"] {
		flags.CryptoKeyPath = fs.Lookup("crypto-key").Value.String()
	}

	if set["config"] {
		flags.ConfigPath = fs.Lookup("config").Value.String()
	} else if set["c"] {
		flags.ConfigPath = fs.Lookup("c").Value.String()
	}
	return flags, nil
}
