package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// messageService is the concrete implementation of the MessageService interface.
type messageService struct {
	messageRepo MessageRepository
	userRepo    UserRepository
	groupRepo   GroupRepository
	hub         Hub
}

// NewMessageService creates a new instance of the MessageService.
func NewMessageService(messageRepo MessageRepository, userRepo UserRepository, groupRepo GroupRepository, hub Hub) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		hub:         hub,
	}
}

// Save implements the MessageService Save method, directly persisting the message.
// This was needed because the MessageService interface expects this method.
func (s *messageService) Save(ctx context.Context, message *Message) (*Message, error) {
	// Proxies the request directly to the repository
	return s.messageRepo.Save(ctx, message)
}

// GetConversationHistory retrieves a list of messages between two users (P2P).
func (s *messageService) GetConversationHistory(ctx context.Context, userID1, userID2 int64, limit int) ([]*Message, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	// FIX: s.messageRepo.FindConversationHistory now matches the interface
	return s.messageRepo.FindConversationHistory(ctx, userID1, userID2, limit) 
}

// SendGroupMessage saves a message to the database and broadcasts it to all group members.
func (s *messageService) SendGroupMessage(ctx context.Context, senderID int64, groupID int64, content string) (*Message, error) {
	// 1. Check if sender is a member of the group
	memberIDs, err := s.groupRepo.FindMembersByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}

	isMember := false
	for _, id := range memberIDs {
		if id == senderID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, errors.New("sender is not a member of this group")
	}

	// 2. Create the message struct
	message := &Message{
		SenderID:    senderID,
		RecipientID: groupID, // Recipient is the Group ID
		Type:        "group",
		Content:     content,
		Timestamp:   time.Now(),
	}

	// 3. Save the message
	// FIX: s.messageRepo.Save now matches the corrected repository interface
	savedMessage, err := s.messageRepo.Save(ctx, message) 
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// 4. Broadcast the message to all group members (WebSocket Hub)
	s.hub.BroadcastGroupMessage(groupID, savedMessage)

	return savedMessage, nil
}
