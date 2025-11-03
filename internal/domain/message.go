package domain

import (
	"context"
	"time"
)

// MessageType defines the type of message (e.g., text, image, system)
type MessageType string

const (
	TextMessage   MessageType = "text"
	ImageMessage  MessageType = "image"
	SystemMessage MessageType = "system"
	TypingMessage MessageType = "typing"
)

// Message is the core data model for a chat message.
type Message struct {
	ID          int64      `json:"id" db:"id"`
	SenderID    int64      `json:"sender_id" db:"sender_id"`
	RecipientID int64       `json:"recipient_id" db:"recipient_id"`
	Type        MessageType `json:"type" db:"type"`
        Content    string       `json:"content" db:"content"`
	MediaURL   string     `json:"media_url" db:"media_url"` 
	Timestamp   time.Time   `json:"timestamp" db:"timestamp"`
}

// ---------------------------------------------
// DEPENDENCY INTERFACES (Needed by the MessageService)
// ---------------------------------------------

// Hub defines the methods the MessageService needs to interact with the WebSocket Hub.
type Hub interface {
	BroadcastGroupMessage(groupID int64, message *Message)
	BroadcastP2PMessage(senderID int64, recipientID int64, message *Message)
}


// MessageRepository defines the data access operations for messages.
type MessageRepository interface {
	Save(ctx context.Context, message *Message) (*Message, error)
	FindConversationHistory(ctx context.Context, userID1, userID2 int64, limit int) ([]*Message, error)
}

// ---------------------------------------------
// SERVICE INTERFACES (Contracts for Business Logic)
// ---------------------------------------------

// MessageService defines the business operations related to messages.
type MessageService interface {
	Save(ctx context.Context, message *Message) (*Message, error)
	GetConversationHistory(ctx context.Context, userID1, userID2 int64, limit int) ([]*Message, error)
	
	// FIX: Update interface signature to match the implementation in message_service.go
	SendGroupMessage(ctx context.Context, senderID int64, groupID int64, content string, mediaURL string) (*Message, error)
	SendP2PMessage(ctx context.Context, senderID int64, recipientID int64, content string, mediaURL string) (*Message, error)

	
}
