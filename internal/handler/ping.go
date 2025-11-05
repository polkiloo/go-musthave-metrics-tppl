package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
)

// Ping handles GET /ping requests by checking database connectivity.
func (h *GinHandler) Ping(c *gin.Context, pool db.Pool) {
	if pool == nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	if c.Request != nil {
		ctx = c.Request.Context()
	}

	err := pool.Ping(ctx)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.Status(http.StatusOK)
}
