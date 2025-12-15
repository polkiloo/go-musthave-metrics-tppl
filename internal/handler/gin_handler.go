package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/audit"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/compression"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/cryptoutil"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/db"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/sign"
)

// GinHandler exposes HTTP handlers that implement the metrics API using Gin.
type GinHandler struct {
	service     service.MetricServiceInterface
	afterUpdate func()
	logger      logger.Logger
	jsonPool    *jsonMetricsPool
}

// NewGinHandler constructs a GinHandler that proxies requests to the provided metric service.
func NewGinHandler(s service.MetricServiceInterface, pool *jsonMetricsPool) *GinHandler {
	if pool == nil {
		pool = NewJSONMetricsPool()
	}
	return &GinHandler{service: s, jsonPool: pool}
}

// RegisterUpdate registers all update endpoints (plain and JSON) on the supplied Gin engine.
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

// RegisterGetValue registers the endpoints that return metric values in both plain text and JSON formats.
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

// RegisterPing registers the database liveness endpoint that responds with HTTP 200 when the pool is ready.
func (h *GinHandler) RegisterPing(r *gin.Engine, pool db.Pool) {
	r.GET("/ping", func(c *gin.Context) {
		h.Ping(c, pool)
	})
}

// RegisterRoutes wires all public handler endpoints to the Gin engine.
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
	R     *gin.Engine
	H     *GinHandler
	L     logger.Logger
	C     compression.Compressor
	S     sign.Signer
	K     sign.SignKey
	A     audit.Publisher      `optional:"true"`
	Clock audit.Clock          `optional:"true"`
	Pool  db.Pool              `optional:"true"`
	D     cryptoutil.Decryptor `optional:"true"`
	TM    gin.HandlerFunc      `name:"trusted-subnet-middleware" optional:"true"`
}) {
	p.H.SetLogger(p.L)
	p.R.Use(logger.Middleware(p.L))
	p.R.Use(cryptoutil.Middleware(p.D))
	p.R.Use(sign.Middleware(p.S, p.K))
	p.R.Use(compression.Middleware(p.C))
	if p.TM != nil {
		p.R.Use(p.TM)
	}
	if p.A != nil {
		p.R.Use(audit.Middleware(p.A, p.L, p.Clock))
	}
	RegisterRoutes(p.R, p.H, p.Pool)
}

// SetAfterUpdateHook installs a callback that is executed after each successful update request.
func (h *GinHandler) SetAfterUpdateHook(fn func()) { h.afterUpdate = fn }

// SetLogger configures the structured logger used by the handler.
func (h *GinHandler) SetLogger(l logger.Logger) { h.logger = l }

// Service returns the underlying MetricServiceInterface used by the handler.
func (h *GinHandler) Service() service.MetricServiceInterface { return h.service }

func (h *GinHandler) jsonMetricsPool() *jsonMetricsPool {
	if h.jsonPool == nil {
		h.jsonPool = NewJSONMetricsPool()
	}
	return h.jsonPool
}

// Module describes the fx module that provides the Gin HTTP handlers.
var Module = fx.Module(
	"handler",
	fx.Provide(
		NewJSONMetricsPool,
		NewGinHandler,
	),
	fx.Invoke(register),
)
