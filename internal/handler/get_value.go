package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
)

func (h *GinHandler) GetValue(c *gin.Context) {
	metricType := models.MetricType(c.Param("type"))
	name := c.Param("name")

	if name == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	v, err := h.service.ProcessGetValue(metricType, name)
	if errors.Is(err, service.ErrUnknownMetricType) {
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

	c.String(http.StatusOK, v)
}

func (h *GinHandler) RegisterValue(r *gin.Engine) {
	r.GET("/value/:type/:name", h.GetValue)
}
