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
	Hub             *Hub
	UserID          int64 // The authenticated ID of the user
	Conn            *websocket.Conn // The actual websocket connection
	Send            chan *Message // Buffered channel of outbound messages
	currentTargetID int64 // ID of the user or group this client is currently talking to
	isGroupChat     bool  // Flag to distinguish between P2P and group chats
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

		// First, try to interpret the payload as a JSON command to set chat context.
		var command struct {
			Action  string `json:"action"`
			UserID  int64  `json:"user_id"`
			GroupID int64  `json:"group_id"`
		}

		if json.Unmarshal(payload, &command) == nil && command.Action == "set_recipient" {
			// It's a command to set the recipient.
			if command.UserID != 0 {
				c.currentTargetID = command.UserID
				c.isGroupChat = false
				log.Printf("User %d set recipient to User %d", c.UserID, c.currentTargetID)
			} else if command.GroupID != 0 {
				c.currentTargetID = command.GroupID
				c.isGroupChat = true
				log.Printf("User %d set recipient to Group %d", c.UserID, c.currentTargetID)
			}
			continue // Command processed, wait for next message.
		}

		// If it's not a command, treat it as a raw text message for the current target.
		if c.currentTargetID == 0 {
			log.Printf("User %d sent a message without setting a recipient. Discarding.", c.UserID)
			// Optional: Send an error back to the client.
			continue
		}

		// Construct the message object from the raw text.
		message := &Message{
			SenderID:  c.UserID,
			Content:   string(payload),
			Timestamp: time.Now(),
			Type:      domain.TextMessage,
		}

		if c.isGroupChat {
			message.GroupID = c.currentTargetID
			message.RecipientID = c.currentTargetID // Hub uses RecipientID for routing
		} else {
			message.RecipientID = c.currentTargetID
		}

		// Broadcast the message through the hub.
		// Note: For features like typing notifications, the client would need to send a structured JSON message.
		c.Hub.Broadcast <- message
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
		Hub:             hub,
		UserID:          userID,
		Conn:            conn,
		Send:            make(chan *Message, 256), // Buffered channel for sending
		currentTargetID: 0, // Initially no target
		isGroupChat:     false,
	}

	// Register the client with the Hub
	client.Hub.Register <- client

	// Start the read and write Goroutines
	go client.writePump()
	go client.readPump()
}
