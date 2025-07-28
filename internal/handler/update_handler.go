package handler

import (
	"net/http"
	"strings"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/service"
)

type UpdateHandler struct {
	service *service.MetricService
}

func NewUpdateHandler() *UpdateHandler {
	service := service.NewMetricService()
	return &UpdateHandler{service: service}
}

func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "text/plain" {
		http.Error(w, "invalid Content-Type, require text/plain", http.StatusBadRequest)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/update/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	typeName, name, raw := parts[0], parts[1], parts[2]

	err := h.service.ProcessUpdate(typeName, name, raw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
