package ws

import (
	"time"

	"github.com/Emmanuel326/chatserver/internal/domain"
)

// Message is the struct used for transport between the client and the Hub.
type Message struct {
	// These IDs should match the domain.User IDs
	ID          int64 `json: "id"`
	SenderID    int64 `json:"sender_id"`
	RecipientID int64 `json:"recipient_id"` // 0 for broadcast/room messages
	Type        domain.MessageType `json:"type"`
	Content     string `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewSystemMessage creates a simple system message for feedback.
func NewSystemMessage(content string) *Message {
	return &Message{
		SenderID:    0, // System user ID
		RecipientID: 0,
		Type:        domain.SystemMessage,
		Content:     content,
		Timestamp:   time.Now(),
	}
}
