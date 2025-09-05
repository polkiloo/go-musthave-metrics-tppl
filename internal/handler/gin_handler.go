package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
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

	r.POST("/updates", func(c *gin.Context) {
		h.UpdatesJSON(c)
	})

	r.POST("/updates/", func(c *gin.Context) {
		h.UpdatesJSON(c)
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

func (h *GinHandler) RegisterPing(r *gin.Engine, pool db.Pool) {
	r.GET("/ping", func(c *gin.Context) {
		h.Ping(c, pool)
	})
}

func RegisterRoutes(r *gin.Engine, h *GinHandler, pool db.Pool) {
	h.RegisterUpdate(r)
	h.RegisterGetValue(r)
	h.RegisterInfo(r)
	if pool != nil {
		h.RegisterPing(r, pool)
	}
}

func register(p struct {
	fx.In
	R    *gin.Engine
	H    *GinHandler
	L    logger.Logger
	C    compression.Compressor
	Pool db.Pool `optional:"true"`
}) {
	p.R.Use(logger.Middleware(p.L), compression.Middleware(p.C))
	RegisterRoutes(p.R, p.H, p.Pool)
}

func (h *GinHandler) SetAfterUpdateHook(fn func()) { h.afterUpdate = fn }

func (h *GinHandler) Service() service.MetricServiceInterface { return h.service }

var Module = fx.Module(
	"handler",
	fx.Provide(NewGinHandler),
	fx.Invoke(register),
)
