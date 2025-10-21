package domain

import (
	"context"
	"time"
)

// messageService is the concrete implementation of the domain.MessageService interface.
type messageService struct {
	messageRepo MessageRepository
}

// NewMessageService creates a new MessageService instance.
func NewMessageService(repo MessageRepository) MessageService {
	return &messageService{messageRepo: repo}
}

// Save handles the persistence of a new message.
func (s *messageService) Save(ctx context.Context, senderID, recipientID int64, msgType MessageType, content string) (*Message, error) {
	
	// Business Rule: Ensure message content is not empty (simple validation)
	if content == "" {
		// Using a specific context error to indicate a business rule violation (or return a custom domain error)
		return nil, context.Canceled 
	}
	
	msg := &Message{
		SenderID: senderID,
		RecipientID: recipientID,
		Type: msgType,
		Content: content,
		Timestamp: time.Now(),
	}
	
	// Delegate persistence to the Repository
	return s.messageRepo.Create(ctx, msg)
}

// GetConversationHistory retrieves message history between two users.
func (s *messageService) GetConversationHistory(ctx context.Context, userID1, userID2 int64, limit int) ([]*Message, error) {
	// Delegate retrieval to the Repository
	return s.messageRepo.FindConversation(ctx, userID1, userID2, limit)
}
