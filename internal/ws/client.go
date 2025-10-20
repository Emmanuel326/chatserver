package ws

import (
	"encoding/json"
	"log"

	"github.com/gofiber/websocket/v2"
)

// Client represents one connected user.
type Client struct {
	Conn *websocket.Conn // The WebSocket connection
	Send chan []byte     // Outgoing messages channel
}

// ReadPump listens for incoming messages from the client.
func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		// Parse incoming message JSON
		var parsed Message
		if err := json.Unmarshal(msg, &parsed); err != nil {
			log.Println("invalid JSON:", err)
			continue
		}

		// Route based on message type
		switch parsed.Type {
		case "text":
			log.Printf("[TEXT] %s: %s\n", parsed.Username, parsed.Content)
			hub.Broadcast <- msg

		case "image":
			log.Printf("[IMAGE] %s sent image: %s\n", parsed.Username, parsed.Content)
			hub.Broadcast <- msg

		default:
			log.Println("unknown message type:", parsed.Type)
		}
	}
}

// WritePump pushes outgoing messages to the client.
func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

