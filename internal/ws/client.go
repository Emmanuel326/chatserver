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
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break
		}

		// Attempt to handle as a JSON command or structured message.
		if c.handleJsonPayload(payload) {
			continue // Handled as JSON, move to the next message.
		}

		// If not JSON, treat as a raw text message.
		c.handleRawTextMessage(payload)
	}
}

// handleJsonPayload attempts to process the payload as a JSON command or message.
// It returns true if the payload was successfully handled as JSON, false otherwise.
func (c *Client) handleJsonPayload(payload []byte) bool {
	var genericMessage map[string]interface{}
	if json.Unmarshal(payload, &genericMessage) != nil {
		return false // Not a valid JSON payload.
	}

	// Case 1: It's a 'set_recipient' command.
	if action, ok := genericMessage["action"].(string); ok && action == "set_recipient" {
		var command struct {
			Action  string `json:"action"`
			UserID  int64  `json:"user_id"`
			GroupID int64  `json:"group_id"`
		}
		_ = json.Unmarshal(payload, &command) // Error can be ignored, already parsed partially.

		if command.UserID != 0 {
			c.currentTargetID = command.UserID
			c.isGroupChat = false
			log.Printf("User %d set recipient to User %d", c.UserID, c.currentTargetID)
		} else if command.GroupID != 0 {
			c.currentTargetID = command.GroupID
			c.isGroupChat = true
			log.Printf("User %d set recipient to Group %d", c.UserID, c.currentTargetID)
		}
		return true
	}

	// Case 2: It's a structured message (e.g., image, typing).
	var message Message
	if err := json.Unmarshal(payload, &message); err == nil {
		if message.Type == "" {
			log.Printf("User %d sent structured message with no type. Discarding.", c.UserID)
			return true // Handled by discarding.
		}

		message.SenderID = c.UserID
		message.Timestamp = time.Now()

		if message.GroupID != 0 {
			message.RecipientID = message.GroupID
		} else if message.RecipientID == 0 {
			if c.currentTargetID == 0 {
				log.Printf("User %d sent structured message with no recipient and no context. Discarding.", c.UserID)
				return true // Handled by discarding.
			}
			message.RecipientID = c.currentTargetID
			if c.isGroupChat {
				message.GroupID = c.currentTargetID
			}
		}

		if message.Type == domain.TypingMessage {
			c.Hub.Typing <- &message
		} else {
			c.Hub.Broadcast <- &message
		}
		return true
	}

	return false // It was JSON but didn't match a known structure.
}

// handleRawTextMessage processes the payload as a plain text message for the current chat context.
func (c *Client) handleRawTextMessage(payload []byte) {
	if c.currentTargetID == 0 {
		log.Printf("User %d sent raw message without setting a recipient. Discarding.", c.UserID)
		return
	}

	message := &Message{
		SenderID:    c.UserID,
		Content:     string(payload),
		Timestamp:   time.Now(),
		Type:        domain.TextMessage,
		RecipientID: c.currentTargetID,
	}
	if c.isGroupChat {
		message.GroupID = c.currentTargetID
	}
	c.Hub.Broadcast <- message
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
