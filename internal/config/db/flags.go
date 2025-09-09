package dbcfg

import (
	"flag"
	"io"
	"os"
)

type DBFlags struct {
	DSN string
}

var (
	defaultDSN = ""
)

func parseFlags() (DBFlags, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.String("d", defaultDSN, "PostreSQL connection string")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return DBFlags{}, err
	}

	flags := DBFlags{}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	if set["d"] {
		flags.DSN = fs.Lookup("d").Value.String()
	}

	return flags, nil
}
