package hub

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/flare19/collabify/internal/protocol"
)

const sendBufferSize = 256

type Client struct {
	id       string
	username string
	room     *Room
	conn     *websocket.Conn
	send     chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()

	for {
		_, rawMsg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("client %s disconnected: %v", c.username, err)
			break
		}

		var msg protocol.Message
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Printf("invalid message from %s: %v", c.username, err)
			continue
		}

		// attach clientId before routing to room
		switch msg.Type {
		case "update", "lang_change":
			c.room.broadcast <- roomMessage{sender: c, msg: msg}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		msg, ok := <-c.send
		if !ok {
			// channel was closed by hub, send close frame
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("write error for %s: %v", c.username, err)
			return
		}
	}
}