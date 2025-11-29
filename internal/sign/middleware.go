package sign

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	defaultBufferSize = 4 << 10   // 4KiB
	maxPooledBuffer   = 256 << 10 // 256KiB
)

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	},
}

func acquireBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func releaseBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	if cap(buf.Bytes()) > maxPooledBuffer {
		return
	}
	buf.Reset()
	bufferPool.Put(buf)
}

// Middleware verifies incoming request signatures and signs outgoing responses when a key is configured.
func Middleware(s Signer, key SignKey) gin.HandlerFunc {
	if key == "" {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		reqBuf := acquireBuffer()
		defer releaseBuffer(reqBuf)

		if _, err := reqBuf.ReadFrom(c.Request.Body); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		body := reqBuf.Bytes()
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Request.ContentLength = int64(len(body))

		if sig := c.GetHeader("HashSHA256"); sig != "" {
			if !s.Verify(body, key, sig) {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		orig := c.Writer
		hw := newSignWriter(orig)
		c.Writer = hw

		c.Next()

		respBody := hw.body.Bytes()
		sig := s.Sign(respBody, key)
		orig.Header().Set("HashSHA256", sig)

		status := hw.status
		if status == 0 {
			status = http.StatusOK
		}
		orig.WriteHeader(status)
		if len(respBody) > 0 {
			_, _ = orig.Write(respBody)
		}
		releaseBuffer(hw.body)
		c.Writer = orig
	}
}

type signWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func newSignWriter(w gin.ResponseWriter) *signWriter {
	return &signWriter{ResponseWriter: w, body: acquireBuffer()}
}

func (w *signWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *signWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

func (w *signWriter) WriteHeader(code int) {
	w.status = code
}

func (w *signWriter) Status() int {
	if w.status != 0 {
		return w.status
	}
	return 0
}

func (w *signWriter) Size() int {
	return w.body.Len()
}

func (w *signWriter) Written() bool {
	return w.body.Len() > 0
}

func (w *signWriter) WriteHeaderNow() {}
