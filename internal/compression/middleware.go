package compression

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Option configures compression middleware behaviour.
type Option func(*settings)
type settings struct {
	allowedCT []string
}

// WithAllowedContentTypes restricts compression to the specified content types.
func WithAllowedContentTypes(ct ...string) Option {
	return func(s *settings) { s.allowedCT = append([]string(nil), ct...) }
}

var defaultCT = []string{"application/json", "text/html"}

// Middleware provides transparent request decompression and response compression for Gin.
func Middleware(cpr Compressor, opts ...Option) gin.HandlerFunc {
	if cpr == nil {
		panic("compression: compressor must not be nil")
	}
	cfg := settings{allowedCT: defaultCT}
	for _, o := range opts {
		o(&cfg)
	}

	return func(c *gin.Context) {
		if enc := c.GetHeader("Content-Encoding"); enc != "" &&
			eqFold(enc, cpr.ContentEncoding()) &&
			isCTAllowed(c.GetHeader("Content-Type"), cfg.allowedCT) {

			rc, err := cpr.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad compressed request body"})
				return
			}
			defer rc.Close()
			c.Request.Body = io.NopCloser(rc)
			c.Request.Header.Del("Content-Encoding")
			c.Request.ContentLength = -1
		}

		if !acceptsEncoding(c.GetHeader("Accept-Encoding"), cpr.ContentEncoding()) {
			c.Next()
			return
		}
		c.Header("Vary", "Accept-Encoding")

		w := newRespWriter(c.Writer, cpr, cfg.allowedCT)
		c.Writer = w
		c.Next()
		w.Close()
	}
}

func eqFold(a, b string) bool {
	if len(a) != len(b) {
		return trimLower(a) == trimLower(b)
	}
	return trimLower(a) == trimLower(b)
}

func trimLower(s string) string {
	i := 0
	j := len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	for i < j && (s[j-1] == ' ' || s[j-1] == '\t') {
		j--
	}
	return strings.ToLower(s[i:j])
}
