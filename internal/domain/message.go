package domain

import (
	"context"
	"time"
)

// MessageType defines the type of message (e.g., text, image, system)
type MessageType string

const (
	TextMessage MessageType = "text"
	ImageMessage MessageType = "image"
	SystemMessage MessageType = "system"
)

// Message is the core data model for a chat message.
type Message struct {
	ID        int64     `json:"id" db:"id"`
	SenderID  int64     `json:"sender_id" db:"sender_id"`
	RecipientID int64   `json:"recipient_id" db:"recipient_id"`
	Type      MessageType `json:"type" db:"type"` 
	Content   string    `json:"content" db:"content"` // Can be text or a file path/URL
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// ---------------------------------------------
// SERVICE INTERFACES (Contracts for Business Logic)
// ---------------------------------------------

// MessageService defines the business operations related to messages.
type MessageService interface {
	Save(ctx context.Context, senderID, recipientID int64, msgType MessageType, content string) (*Message, error)
	GetConversationHistory(ctx context.Context, userID1, userID2 int64, limit int) ([]*Message, error)
}

// MessageRepository defines the data access operations for messages.
type MessageRepository interface {
	Create(ctx context.Context, message *Message) (*Message, error)
	FindConversation(ctx context.Context, userID1, userID2 int64, limit int) ([]*Message, error)
}
