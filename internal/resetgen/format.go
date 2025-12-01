package resetgen

import (
	"go/format"
	"os"
)

var (
	formatSourceFunc = formatSource
	writeFileFunc    = writeFile
)

func formatSource(src []byte) ([]byte, error) {
	return format.Source(src)
}

func writeFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0o644)
}
