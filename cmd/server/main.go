package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
)

func main() {
	r := gin.Default()
	h := handler.NewGinHandler()
	handler.RegisterRoutes(r, h)

	addr := ":8080"
	log.Printf("Server listening on http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
