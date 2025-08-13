package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	models "github.com/polkiloo/go-musthave-metrics-tppl/internal/model"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
)

func (h *GinHandler) Update(c *gin.Context) {
	metricType := models.MetricType(c.Param("type"))
	name := c.Param("name")
	raw := c.Param("value")

	if name == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := h.service.ProcessUpdate(metricType, name, raw)
	if errors.Is(err, service.ErrUnknownMetricType) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.String(http.StatusOK, "ok")
}

func (h *GinHandler) RegisterUpdate(r *gin.Engine) {
	r.POST("/update/:type/:name/:value", h.Update)
}
