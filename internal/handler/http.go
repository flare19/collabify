package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type HTTPHandler struct{}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

func (h *HTTPHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	roomID := uuid.NewString()[:8]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"roomId": roomID,
	})
}

func (h *HTTPHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/"+r.URL.Path[1:])
}