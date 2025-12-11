package cryptoutil

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
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
