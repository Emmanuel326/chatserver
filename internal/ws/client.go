package ws

import (
	"encoding/json" // <-- ADDED: For JSON deserialization
	"log"
	"time"

	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/gorilla/websocket"
)

// Define buffer/timeout settings
const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the Hub.
type Client struct {
	Hub *Hub
	UserID int64 // The authenticated ID of the user
	Conn *websocket.Conn // The actual websocket connection
	Send chan *Message // Buffered channel of outbound messages
}

// readPump pumps messages from the websocket connection to the Hub.
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(maxMessageSize)
	// Set an initial read deadline to ensure the connection isn't idle forever
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	// Set a handler for pong messages to reset the read deadline
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait));
		return nil
	})

	for {
		// Reads the message type and payload
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break
		}
		
		// Unmarshal the incoming JSON message
		var message Message
		if err := json.Unmarshal(payload, &message); err != nil {
			log.Printf("Error unmarshalling message from UserID %d: %v", c.UserID, err)
			continue
		}

		// Populate server-side fields for security and consistency
		message.SenderID = c.UserID
		message.Timestamp = time.Now()

		// If GroupID is present, treat it as the recipient for routing in the hub.
		// The domain layer uses RecipientID for both P2P and group messages.
		if message.GroupID != 0 {
			message.RecipientID = message.GroupID
		}

		// Route message to appropriate hub channel based on its type
		if message.Type == domain.TypingMessage {
			// Typing notifications are transient and not persisted
			c.Hub.Typing <- &message
		} else {
			// Other messages are persisted and broadcast
			c.Hub.Broadcast <- &message
		}
	}
}

// writePump pumps messages from the Hub's Send channel to the websocket connection.
func (c *Client) writePump() {
	// Ticker sends ping messages periodically to keep the connection alive
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// Set a write deadline
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The Hub closed the channel (unregistered the client)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			// Write the message to the client
			err := c.Conn.WriteJSON(message) // Write as JSON for structured output
			if err != nil {
				return
			}

		case <-ticker.C:
			// Send a Ping message
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // Write failed, stop the pump
			}
		}
	}
}

// ServeWs handles the websocket request from the peer.
func ServeWs(hub *Hub, conn *websocket.Conn, userID int64) {
	client := &Client{
		Hub: hub,
		UserID: userID,
		Conn: conn,
		Send: make(chan *Message, 256), // Buffered channel for sending
	}
	
	// Register the client with the Hub
	client.Hub.Register <- client

	// Start the read and write Goroutines
	go client.writePump()
	go client.readPump()
}
