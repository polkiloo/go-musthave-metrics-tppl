package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/storage"
)

func Example_plainEndpoints() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	store := storage.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewGinHandler(svc, handler.NewJSONMetricsPool())
	h.RegisterUpdate(router)
	h.RegisterGetValue(router)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/42.5", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	fmt.Println("update status:", w.Code)

	req = httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", http.NoBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	fmt.Println("value status:", w.Code)
	fmt.Println("body:", strings.TrimSpace(w.Body.String()))

	// Output:
	// update status: 200
	// value status: 200
	// body: 42.5
}

func Example_jsonEndpoints() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	store := storage.NewMemStorage()
	svc := service.NewMetricService(store)
	h := handler.NewGinHandler(svc, handler.NewJSONMetricsPool())
	h.RegisterUpdate(router)
	h.RegisterGetValue(router)

	value := 100.0
	payload := models.Metrics{ID: "Alloc", MType: models.GaugeType, Value: &value}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	fmt.Println("update status:", w.Code)
	fmt.Println("update body:", strings.TrimSpace(w.Body.String()))

	query := models.Metrics{ID: "Alloc", MType: models.GaugeType}
	queryBody, _ := json.Marshal(query)
	req = httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(queryBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	fmt.Println("value status:", w.Code)
	fmt.Println("value body:", strings.TrimSpace(w.Body.String()))

	// Output:
	// update status: 200
	// update body: {"id":"Alloc","type":"gauge","value":100}
	// value status: 200
	// value body: {"id":"Alloc","type":"gauge","value":100}
}
