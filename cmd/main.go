package main

import (
	"log"
	"net/http"

	"github.com/flare19/collabify/internal/handler"
	"github.com/flare19/collabify/internal/hub"
)

func main() {
	h := hub.NewHub()
	go h.Run()

	wsHandler := handler.NewWSHandler(h)
	httpHandler := handler.NewHTTPHandler()

	mux := http.NewServeMux()

	// static files
	mux.HandleFunc("GET /", httpHandler.ServeStatic)

	// rest endpoints
	mux.HandleFunc("POST /rooms", httpHandler.CreateRoom)

	// websocket
	mux.HandleFunc("GET /ws/{roomId}", wsHandler.ServeWS)

	log.Println("server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}