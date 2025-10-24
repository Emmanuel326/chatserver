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
	ID          int64       `json:"id" db:"id"`
	SenderID    int64       `json:"sender_id" db:"sender_id"`
	RecipientID int64       `json:"recipient_id" db:"recipient_id"`
	Type        MessageType `json:"type" db:"type"`
	Content     string      `json:"content" db:"content"` // Can be text or a file path/URL
	Timestamp   time.Time   `json:"timestamp" db:"timestamp"`
}

// ---------------------------------------------
// DEPENDENCY INTERFACES (Needed by the MessageService)
// ---------------------------------------------

// Hub defines the methods the MessageService needs to interact with the WebSocket Hub.
// This fixes the "undefined: Hub" error in message_service.go
type Hub interface {
    BroadcastGroupMessage(groupID int64, message *Message)
}


// MessageRepository defines the data access operations for messages.
type MessageRepository interface {
	// FIX: Rename Create to Save to match implementation usage (and assumed intent)
	Save(ctx context.Context, message *Message) (*Message, error)
	// FIX: Rename FindConversation to FindConversationHistory to match service usage
	FindConversationHistory(ctx context.Context, userID1, userID2 int64, limit int) ([]*Message, error)
	
	// REMOVE: Repository should not have service methods like SendGroupMessage
}

// ---------------------------------------------
// SERVICE INTERFACES (Contracts for Business Logic)
// ---------------------------------------------

// MessageService defines the business operations related to messages.
type MessageService interface {
	// FIX: Update Save to take Message struct, not raw fields (consistency)
	Save(ctx context.Context, message *Message) (*Message, error)
	GetConversationHistory(ctx context.Context, userID1, userID2 int64, limit int) ([]*Message, error)
	
	// ADD: The method needed by the API handler
	SendGroupMessage(ctx context.Context, senderID int64, groupID int64, content string) (*Message, error)
}
