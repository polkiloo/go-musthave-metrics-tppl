package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Info responds with a minimal HTML page for the service root.
func (h *GinHandler) Info(c *gin.Context) {
	c.Data(http.StatusOK, gin.MIMEHTML, []byte("<html><body></body></html>"))
}

// RegisterInfo registers the root informational endpoint.
func (h *GinHandler) RegisterInfo(r *gin.Engine) {
	r.GET("/", h.Info)
}
