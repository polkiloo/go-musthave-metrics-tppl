package handler

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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

	// вызов функции, которую мы тестируем
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
		fx.NopLogger,
		// Предоставляем *gin.Engine для invoke(register)
		fx.Provide(func() *gin.Engine { return gin.New() }),
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
		// Провайдим *gin.Engine, чтобы register смог подвязаться
		fx.Provide(func() *gin.Engine { return gin.New() }),
		Module,
		// Достаём из контейнера тот же *gin.Engine для проверок
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

	// После запуска register уже вызван, маршруты должны быть на месте
	if !hasRoute(router, "POST", "/update/:type/:name/:value") {
		t.Fatalf("expected route POST /update/:type/:name/:value")
	}
	if !hasRoute(router, "GET", "/value/:type/:name") {
		t.Fatalf("expected route GET /value/:type/:name")
	}
}
