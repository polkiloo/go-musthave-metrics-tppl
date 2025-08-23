package compression

import (
	"io"
	"strings"
)

type Compressor interface {
	ContentEncoding() string
	NewWriter(w io.Writer) (io.WriteCloser, error)
	NewReader(r io.Reader) (io.ReadCloser, error)
}

func isCTAllowed(ct string, allowed []string) bool {
	if ct == "" {
		return false
	}
	ct = strings.ToLower(ct)
	for _, p := range allowed {
		if strings.HasPrefix(ct, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func acceptsEncoding(ae, enc string) bool {
	ae = strings.ToLower(ae)
	enc = strings.ToLower(enc)
	if ae == "" {
		return false
	}
	for _, tok := range strings.Split(ae, ",") {
		tok = strings.TrimSpace(tok)
		if i := strings.Index(tok, ";"); i >= 0 {
			tok = tok[:i]
		}
		if tok == enc || tok == "*" {
			return true
		}
	}
	return false
}
