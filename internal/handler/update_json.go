package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/audit"
)

// UpdateJSON handles POST /update requests that transmit metrics in JSON format.
func (h *GinHandler) UpdateJSON(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	}

	pool := h.jsonMetricsPool()
	in := pool.AcquireMetric()
	defer pool.ReleaseMetric(in)

	if err := json.NewDecoder(c.Request.Body).Decode(in); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if in.ID == "" || in.MType == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := h.service.ProcessUpdate(in)

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	audit.AddRequestMetrics(c, in.ID)

	if h.afterUpdate != nil {
		h.afterUpdate()
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, in)
}
