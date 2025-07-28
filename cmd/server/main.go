package main

import (
	"log"
	"net/http"

	"github.com/polkiloo/go-musthave-metrics-tppl/internal/handler"
)

func main() {

	h := handler.NewUpdateHandler()

	http.HandleFunc("/update/", h.Handle)
	addr := ":8080"
	log.Printf("Server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
