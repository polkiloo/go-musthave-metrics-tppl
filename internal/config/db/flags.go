package dbcfg

import (
	"flag"
	"io"
	"os"
)

type DBFlags struct {
	DSN        string
	ConfigPath string
}

var (
	defaultDSN = ""
)

func parseFlags() (DBFlags, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.String("d", defaultDSN, "PostreSQL connection string")
	fs.String("c", "", "path to configuration file")
	fs.String("config", "", "path to configuration file")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return DBFlags{}, err
	}

	flags := DBFlags{}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	if set["d"] {
		flags.DSN = fs.Lookup("d").Value.String()
	}

	if set["config"] {
		flags.ConfigPath = fs.Lookup("config").Value.String()
	} else if set["c"] {
		flags.ConfigPath = fs.Lookup("c").Value.String()
	}

	return flags, nil
}
