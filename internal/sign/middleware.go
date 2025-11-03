package sign

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Middleware(s Signer, key SignKey) gin.HandlerFunc {
	if key == "" {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		if sig := c.GetHeader("HashSHA256"); sig != "" {
			if !s.Verify(body, key, sig) {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		orig := c.Writer
		hw := &signWriter{ResponseWriter: orig}
		c.Writer = hw

		c.Next()

		sig := s.Sign(hw.body.Bytes(), key)
		orig.Header().Set("HashSHA256", sig)

		status := hw.status
		if status == 0 {
			status = http.StatusOK
		}
		orig.WriteHeader(status)
		if hw.body.Len() > 0 {
			_, _ = orig.Write(hw.body.Bytes())
		}
		c.Writer = orig
	}
}

type signWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
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
