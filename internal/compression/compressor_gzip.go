package compression

import (
	"compress/gzip"
	"io"

	"go.uber.org/fx"
)

const (
	// DefaultCompression mirrors gzip.DefaultCompression for convenience.
	DefaultCompression = gzip.DefaultCompression
	// BestSpeed mirrors gzip.BestSpeed for convenience.
	BestSpeed = gzip.BestSpeed
	// BestCompression mirrors gzip.BestCompression for convenience.
	BestCompression = gzip.BestCompression
)

// Gzip implements Compressor using the gzip algorithm.
type Gzip struct{ level int }

// NewGzip constructs a Gzip compressor with the specified level.
func NewGzip(level int) *Gzip { return &Gzip{level: level} }

// ContentEncoding returns the HTTP Content-Encoding token for gzip.
func (g *Gzip) ContentEncoding() string { return "gzip" }

// NewWriter creates a gzip writer using the configured compression level.
func (g *Gzip) NewWriter(w io.Writer) (io.WriteCloser, error) { return gzip.NewWriterLevel(w, g.level) }

// NewReader creates a gzip reader for compressed payloads.
func (g *Gzip) NewReader(r io.Reader) (io.ReadCloser, error) { return gzip.NewReader(r) }

// Module registers the default gzip compressor in the fx container.
var Module = fx.Module(
	"compressor",
	fx.Provide(func() Compressor { return NewGzip(gzip.BestSpeed) }),
)
