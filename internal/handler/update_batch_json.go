package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/audit"
)

// UpdatesJSON handles POST /updates requests that submit batches of metrics in JSON format.
func (h *GinHandler) UpdatesJSON(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	}
	batch := acquireMetricsBatch()
	defer releaseMetricsBatch(batch)

	metrics := *batch
	if err := json.NewDecoder(c.Request.Body).Decode(&metrics); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	*batch = metrics
	metrics = *batch
	for _, m := range metrics {
		if m.ID == "" || m.MType == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	if err := h.service.ProcessUpdates(metrics); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	for i := range metrics {
		audit.AddRequestMetrics(c, metrics[i].ID)
	}

	if h.afterUpdate != nil {
		h.afterUpdate()
	}
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, metrics)
}
