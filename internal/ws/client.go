package ws

import (
	"encoding/json" // <-- ADDED: For JSON deserialization
	"log"
	"time"

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
		
		// --- FIX START ---
		var message Message // Use the Message struct from models.go
		if err := json.Unmarshal(payload, &message); err != nil {
            log.Printf("JSON unmarshal error from UserID %d: %v", c.UserID, err)
            continue // Skip invalid message
        }

		// CRITICAL FIX: Inject the sender's ID (the only authenticated ID)
		message.SenderID = c.UserID 
        
        // Add server-side timestamp
        message.Timestamp = time.Now()
		// --- FIX END ---

		// Send the message to the Hub's broadcast channel
		c.Hub.Broadcast <- &message // Send the pointer to the message
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
