package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
)

// GetValueJSON handles POST /value requests returning the metric value as JSON.
func (h *GinHandler) GetValueJSON(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	}

	var q models.Metrics
	if err := json.NewDecoder(c.Request.Body).Decode(&q); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if q.ID == "" || q.MType == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	metric, err := h.service.ProcessGetValue(q.ID, q.MType)
	switch {
	case errors.Is(err, service.ErrMetricNotFound), errors.Is(err, models.ErrMetricUnknownName):
		c.AbortWithStatus(http.StatusNotFound)
		return
	case err != nil:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, metric)
}
