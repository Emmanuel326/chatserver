
// ---------------------------------------------
// Represents a single WebSocket client connection.
// Handles sending and receiving messages from that client.
// ---------------------------------------------

package ws

import (
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
			break
		}
		hub.Broadcast <- msg // Send to all clients via hub
	}
}

// WritePump pushes outgoing messages to the client.
func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

