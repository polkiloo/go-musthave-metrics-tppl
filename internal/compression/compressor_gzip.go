package compression

import (
	"compress/gzip"
	"io"

	"go.uber.org/fx"
)

const (
	DefaultCompression = gzip.DefaultCompression
	BestSpeed          = gzip.BestSpeed
	BestCompression    = gzip.BestCompression
)

type Gzip struct{ level int }

func NewGzip(level int) *Gzip { return &Gzip{level: level} }

func (g *Gzip) ContentEncoding() string                       { return "gzip" }
func (g *Gzip) NewWriter(w io.Writer) (io.WriteCloser, error) { return gzip.NewWriterLevel(w, g.level) }
func (g *Gzip) NewReader(r io.Reader) (io.ReadCloser, error)  { return gzip.NewReader(r) }

var Module = fx.Module(
	"compressor",
	fx.Provide(func() Compressor { return NewGzip(gzip.BestSpeed) }),
)
