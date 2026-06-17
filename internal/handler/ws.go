package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/flare19/collabify/internal/hub"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // fine for dev, lock this down in production
	},
}

type WSHandler struct {
	hub *hub.Hub
}

func NewWSHandler(h *hub.Hub) *WSHandler {
	return &WSHandler{hub: h}
}

func (wsh *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomId")
	username := r.URL.Query().Get("username")

	if roomID == "" || username == "" {
		http.Error(w, "roomId and username are required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	wsh.hub.JoinRoom(roomID, username, conn)
}