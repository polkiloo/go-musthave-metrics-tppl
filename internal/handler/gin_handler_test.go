package handler

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/logger"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
	"go.uber.org/fx"
)

func hasRoute(r *gin.Engine, method, path string) bool {
	for _, ri := range r.Routes() {
		if ri.Method == method && ri.Path == path {
			return true
		}
	}
	return false
}

func TestNewGinHandler_ServiceNotNil(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewGinHandler()
	if h == nil {
		t.Fatalf("NewGinHandler returned nil")
	}
	if h.service == nil {
		t.Fatalf("handler.service is nil")
	}
}

func TestRegisterRoutes_RegistersExpectedEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	h := NewGinHandler()

	RegisterRoutes(r, h)

	if !hasRoute(r, "POST", "/update/:type/:name/:value") {
		t.Fatalf("expected route POST /update/:type/:name/:value to be registered; got: %+v", r.Routes())
	}
	if !hasRoute(r, "GET", "/value/:type/:name") {
		t.Fatalf("expected route GET /value/:type/:name to be registered; got: %+v", r.Routes())
	}
}

func TestModule_WiringIsValid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	err := fx.ValidateApp(
		fx.Provide(func() *gin.Engine { return gin.New() }),
		fx.Supply(
			fx.Annotate(&test.FakeLogger{}, fx.As(new(logger.Logger))),
		),

		Module,
	)
	if err != nil {
		t.Fatalf("fx wiring validation failed: %v", err)
	}
}

func TestModule_StartsAndRegistersRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var router *gin.Engine

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() *gin.Engine { return gin.New() }),
		fx.Supply(
			fx.Annotate(&test.FakeLogger{}, fx.As(new(logger.Logger))),
		),
		Module,
		fx.Populate(&router),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("fx.Start: %v", err)
	}
	defer func() { _ = app.Stop(ctx) }()

	if router == nil {
		t.Fatalf("router was not populated")
	}

	if !hasRoute(router, "POST", "/update/:type/:name/:value") {
		t.Fatalf("expected route POST /update/:type/:name/:value")
	}
	if !hasRoute(router, "GET", "/value/:type/:name") {
		t.Fatalf("expected route GET /value/:type/:name")
	}
}
