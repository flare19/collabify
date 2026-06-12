package hub

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type registerRequest struct {
	roomID   string
	username string
	conn     *websocket.Conn
}

type Hub struct {
	rooms      map[string]*Room
	mu         sync.RWMutex
	register   chan registerRequest
	unregister chan *Room
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan registerRequest),
		unregister: make(chan *Room),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case req := <-h.register:
			h.mu.Lock()
			room, exists := h.rooms[req.roomID]
			if !exists {
				room = NewRoom(req.roomID, "javascript")
				h.rooms[req.roomID] = room
				go room.Run()
				log.Printf("room created: %s", req.roomID)
			}
			h.mu.Unlock()

			client := &Client{
				id:       uuid.NewString(),
				username: req.username,
				room:     room,
				conn:     req.conn,
				send:     make(chan []byte, sendBufferSize),
			}

			room.mu.Lock()
			room.clients[client.id] = client
			room.mu.Unlock()

			room.sendSync(client)
			room.broadcastPresence()

			go client.readPump()
			go client.writePump()

			log.Printf("client %s joined room %s", req.username, req.roomID)

		case room := <-h.unregister:
			h.mu.Lock()
			if _, exists := h.rooms[room.id]; exists {
				delete(h.rooms, room.id)
				log.Printf("room deleted: %s", room.id)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) JoinRoom(roomID string, username string, conn *websocket.Conn) {
	h.register <- registerRequest{
		roomID:   roomID,
		username: username,
		conn:     conn,
	}
}