package logger

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware logs basic request and response information using the provided logger.
func Middleware(l Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		if size < 0 {
			size = 0
		}

		l.WriteInfo("request",
			"method", c.Request.Method,
			"uri", c.Request.RequestURI,
			"duration", duration,
		)

		l.WriteInfo("response",
			"status", status,
			"size", size,
		)
	}
}
