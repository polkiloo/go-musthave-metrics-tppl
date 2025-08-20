package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *GinHandler) Info(c *gin.Context) {
	c.Data(http.StatusOK, gin.MIMEHTML, []byte("<html><body></body></html>"))
}

func (h *GinHandler) RegisterInfo(r *gin.Engine) {
	r.GET("/", h.Info)
}
