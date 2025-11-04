package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/audit"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

func (h *GinHandler) UpdatePlain(c *gin.Context) {
	metricType := models.MetricType(c.Param("type"))
	if !metricType.IsValid() {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	metricName := c.Param("name")
	if metricName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	m, err := models.NewMetrics(metricName, c.Param("value"), metricType)
	if errors.Is(err, models.ErrMetricUnknownName) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = h.service.ProcessUpdate(m)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	audit.AddRequestMetrics(c, metricName)

	if h.afterUpdate != nil {
		h.afterUpdate()
	}

	c.String(http.StatusOK, "ok")
}
