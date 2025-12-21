package trustedsubnet

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
	"go.uber.org/fx"
)

const middlewareName = "trusted-subnet-middleware"

// shouldValidate indicates whether the request must be validated against the trusted subnet.
func shouldValidate(r *http.Request) bool {
	if r == nil || r.URL == nil {
		return false
	}
	if r.Method != http.MethodPost {
		return false
	}
	path := r.URL.Path
	return strings.HasPrefix(path, "/update") || strings.HasPrefix(path, "/updates")
}

// NewMiddleware builds a Gin middleware that validates X-Real-IP against the configured trusted subnet.
// When the subnet is empty, no middleware is returned and validation is skipped.
func NewMiddleware(cfg *server.AppConfig, l logger.Logger) (gin.HandlerFunc, error) {
	validator, err := newValidator(cfg)
	if err != nil {
		return nil, err
	}

	if validator == nil {
		return nil, nil
	}

	return func(c *gin.Context) {
		if !shouldValidate(c.Request) {
			c.Next()
			return
		}

		if !validator.contains(c.Request.Header.Get("X-Real-IP")) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}, nil
}

// Module wires the trusted subnet middleware into the fx container.
var Module = fx.Module("trustedsubnet",
	fx.Provide(fx.Annotate(NewMiddleware, fx.ResultTags(`name:"`+middlewareName+`"`))),
)

// Name returns the fx name used for the middleware, simplifying injections.
func Name() string {
	return middlewareName
}
