package cryptoutil

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxEncryptedBodySize = 10 << 20

// Middleware decrypts incoming requests using the provided Decryptor.
func Middleware(dec Decryptor) gin.HandlerFunc {
	return func(c *gin.Context) {
		encryptedKey := c.GetHeader(CryptoKeyHeader)
		if encryptedKey == "" {
			c.Next()
			return
		}
		if dec == nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		var bodyReader io.Reader
		if c.Request.ContentLength < 0 || c.Request.ContentLength > maxEncryptedBodySize {
			bodyReader = http.MaxBytesReader(c.Writer, c.Request.Body, maxEncryptedBodySize)
		} else {
			bodyReader = io.LimitReader(c.Request.Body, c.Request.ContentLength)
		}

		body, err := io.ReadAll(bodyReader)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				c.AbortWithStatus(http.StatusRequestEntityTooLarge)
				return
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		plain, err := dec.Decrypt(body, encryptedKey)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(plain))
		c.Request.ContentLength = int64(len(plain))
		c.Request.Header.Del(CryptoKeyHeader)
		c.Next()
	}
}
