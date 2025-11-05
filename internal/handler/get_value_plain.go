package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
)

// GetValuePlain handles GET /value/:type/:name requests returning metric values as plain text.
func (h *GinHandler) GetValuePlain(c *gin.Context) {
	metricType := models.MetricType(c.Param("type"))
	if !metricType.IsValid() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	metricName := c.Param("name")
	if metricName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	v, err := h.service.ProcessGetValue(metricName, metricType)
	if errors.Is(err, models.ErrMetricInvalidType) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if errors.Is(err, service.ErrMetricNotFound) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch v.MType {
	case models.GaugeType:
		if v.Value == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, strconv.FormatFloat(*v.Value, 'f', -1, 64))
	case models.CounterType:
		if v.Delta == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, strconv.FormatInt(*v.Delta, 10))
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

}
