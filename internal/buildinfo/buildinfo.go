package buildinfo

import (
	"fmt"
	"io"
	"strings"

	_ "embed"
)

//go:embed .version
var versionFile string

const defaultValue = "N/A"

// Info contains build metadata values.
type Info struct {
	Version string
	Date    string
	Commit  string
}

var info = parseVersionFile(versionFile)

// InfoData returns embedded build information parsed from the generated version file.
func InfoData() Info {
	return info
}

// Print writes build information to the provided writer.
func Print(w io.Writer, info Info) {
	fmt.Fprintf(w, "Build version: %s\n", valueOrNA(info.Version))
	fmt.Fprintf(w, "Build date: %s\n", valueOrNA(info.Date))
	fmt.Fprintf(w, "Build commit: %s\n", valueOrNA(info.Commit))
}

func parseVersionFile(contents string) Info {
	info := Info{}

	lines := strings.Split(contents, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "version":
			info.Version = val
		case "date":
			info.Date = val
		case "commit":
			info.Commit = val
		}
	}

	return info
}

func valueOrNA(value string) string {
	if value == "" {
		return defaultValue
	}

	return value
}
