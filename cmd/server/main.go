package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
)

func main() {
	addr, err := parseFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	r := gin.Default()
	h := handler.NewGinHandler()
	handler.RegisterRoutes(r, h)

	log.Printf("Server listening on http://%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
