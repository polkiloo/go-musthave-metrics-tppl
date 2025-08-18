package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/models"
)

func (h *GinHandler) UpdateJSON(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
		c.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	}

	var in models.Metrics
	if err := json.NewDecoder(c.Request.Body).Decode(&in); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if in.ID == "" || in.MType == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := h.service.ProcessUpdate(&in)

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, in)
}
