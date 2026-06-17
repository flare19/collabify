package hub

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/flare19/collabify/internal/protocol"
)

type roomMessage struct {
	sender *Client
	msg    protocol.Message
}

type Room struct {
	id         string
	clients    map[string]*Client
	doc        string
	version    int
	language   string
	mu         sync.RWMutex
	broadcast  chan roomMessage
	unregister chan *Client
}

func NewRoom(id string, language string) *Room {
	return &Room{
		id:         id,
		clients:    make(map[string]*Client),
		doc:        "",
		version:    0,
		language:   language,
		broadcast:  make(chan roomMessage, 256),
		unregister: make(chan *Client),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.unregister:
			if _, ok := r.clients[client.id]; ok {
				delete(r.clients, client.id)
				close(client.send)
				r.broadcastPresence()
			}

		case rm := <-r.broadcast:
			switch rm.msg.Type {
			case "update":
				r.handleUpdate(rm)
			case "lang_change":
				r.handleLangChange(rm)
			}
		}
	}
}

func (r *Room) handleUpdate(rm roomMessage) {
	var payload protocol.UpdatePayload
	if err := json.Unmarshal(rm.msg.Payload, &payload); err != nil {
		log.Printf("invalid update payload: %v", err)
		return
	}

	r.mu.Lock()
	r.version++
	payload.Version = r.version
	payload.ClientID = rm.sender.id
	r.mu.Unlock()

	updatedPayload, _ := json.Marshal(payload)
	outMsg, _ := json.Marshal(protocol.Message{
		Type:    "update",
		Payload: updatedPayload,
	})

	for _, client := range r.clients {
		client.send <- outMsg
	}
}

func (r *Room) handleLangChange(rm roomMessage) {
	var payload protocol.LangChangePayload
	if err := json.Unmarshal(rm.msg.Payload, &payload); err != nil {
		log.Printf("invalid lang_change payload: %v", err)
		return
	}

	r.mu.Lock()
	r.language = payload.Language
	r.mu.Unlock()

	outMsg, _ := json.Marshal(protocol.Message{
		Type:    "lang_change",
		Payload: rm.msg.Payload,
	})

	for _, client := range r.clients {
		client.send <- outMsg
	}
}

func (r *Room) broadcastPresence() {
	users := make([]string, 0, len(r.clients))
	for _, c := range r.clients {
		users = append(users, c.username)
	}

	payload, _ := json.Marshal(protocol.PresencePayload{Users: users})
	outMsg, _ := json.Marshal(protocol.Message{
		Type:    "presence",
		Payload: payload,
	})

	for _, client := range r.clients {
		client.send <- outMsg
	}
}

func (r *Room) sendSync(client *Client) {
	r.mu.RLock()
	doc := r.doc
	version := r.version
	language := r.language
	r.mu.RUnlock()

	users := make([]string, 0, len(r.clients))
	for _, c := range r.clients {
		users = append(users, c.username)
	}

	payload, _ := json.Marshal(protocol.SyncPayload{
		Doc:      doc,
		Version:  version,
		Language: language,
		Users:    users,
	})

	outMsg, _ := json.Marshal(protocol.Message{
		Type:    "sync",
		Payload: payload,
	})

	client.send <- outMsg
}