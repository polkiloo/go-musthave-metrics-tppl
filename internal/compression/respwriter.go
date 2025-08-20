package compression

import (
	"io"

	"github.com/gin-gonic/gin"
)

type respWriter struct {
	gin.ResponseWriter
	cpr       Compressor
	allowedCT []string

	wc         io.WriteCloser
	using      bool
	headerSent bool
}

func newRespWriter(w gin.ResponseWriter, cpr Compressor, allowed []string) *respWriter {
	return &respWriter{ResponseWriter: w, cpr: cpr, allowedCT: allowed}
}

func (w *respWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *respWriter) Close() {
	if w.using && w.wc != nil {
		_ = w.wc.Close()
	}
}

func (w *respWriter) WriteHeaderNow() {
	w.ResponseWriter.WriteHeaderNow()
}

func (w *respWriter) Write(b []byte) (int, error) {
	if !w.headerSent {
		w.headerSent = true

		h := w.Header()

		if h.Get("Content-Encoding") == "" {
			if ct := h.Get("Content-Type"); ct != "" && isCTAllowed(ct, w.allowedCT) {
				h.Del("Content-Length")
				h.Set("Content-Encoding", w.cpr.ContentEncoding())

				wc, err := w.cpr.NewWriter(w.ResponseWriter)
				if err == nil {
					w.wc = wc
					w.using = true
				} else {
					h.Del("Content-Encoding")
				}
			}
		}
	}

	if w.using {
		return w.wc.Write(b)
	}
	return w.ResponseWriter.Write(b)
}
