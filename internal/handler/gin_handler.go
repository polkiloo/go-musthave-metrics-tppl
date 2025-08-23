package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
)

type GinHandler struct {
	service     service.MetricServiceInterface
	afterUpdate func()
}

func NewGinHandler(s service.MetricServiceInterface) *GinHandler {
	return &GinHandler{service: s}
}

func (h *GinHandler) RegisterUpdate(r *gin.Engine) {
	r.POST("/update", func(c *gin.Context) {
		h.UpdateJSON(c)
	})

	r.POST("/update/", func(c *gin.Context) {
		h.UpdateJSON(c)
	})

	r.POST("/update/:type/:name/:value", func(c *gin.Context) {
		h.UpdatePlain(c)
	})
}

func (h *GinHandler) RegisterGetValue(r *gin.Engine) {
	r.POST("/value", func(c *gin.Context) {
		h.GetValueJSON(c)
	})

	r.POST("/value/", func(c *gin.Context) {
		h.GetValueJSON(c)
	})

	r.GET("/value/:type/:name", func(c *gin.Context) {
		h.GetValuePlain(c)
	})
}

func RegisterRoutes(r *gin.Engine, h *GinHandler) {
	h.RegisterUpdate(r)
	h.RegisterGetValue(r)
	h.RegisterInfo(r)
}

func register(r *gin.Engine, h *GinHandler, l logger.Logger, c compression.Compressor) {
	r.Use(logger.Middleware(l), compression.Middleware(c))
	RegisterRoutes(r, h)
}

func (h *GinHandler) SetAfterUpdateHook(fn func()) { h.afterUpdate = fn }

func (h *GinHandler) Service() service.MetricServiceInterface { return h.service }

var Module = fx.Module(
	"handler",
	fx.Provide(func() service.MetricServiceInterface { return service.NewMetricService() }, NewGinHandler),
	fx.Invoke(register),
)
