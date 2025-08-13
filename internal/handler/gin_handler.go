package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
)

type GinHandler struct {
	service service.MetricServiceInterface
}

func NewGinHandler() *GinHandler {
	return &GinHandler{service: service.NewMetricService()}
}

func RegisterRoutes(r *gin.Engine, h *GinHandler) {
	h.RegisterUpdate(r)
	h.RegisterValue(r)
}
