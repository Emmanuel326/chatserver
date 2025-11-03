package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// messageService is the concrete implementation of the MessageService interface.
type messageService struct {
	messageRepo  MessageRepository
	userRepo   UserRepository
	groupRepo   GroupRepository
	hub        Hub
}

// NewMessageService creates a new instance of the MessageService.
func NewMessageService(messageRepo MessageRepository, userRepo UserRepository, groupRepo GroupRepository, hub Hub) MessageService {
	return &messageService{
		messageRepo:  messageRepo,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		hub:           hub,
	}
}

// Save implements the MessageService Save method, directly persisting the message.
func (s *messageService) Save(ctx context.Context, message *Message) (*Message, error) {
	return s.messageRepo.Save(ctx, message)
}

// GetConversationHistory retrieves a list of messages between two users (P2P).
func (s *messageService) GetConversationHistory(ctx context.Context, userID1, userID2 int64, limit int, beforeID int64) ([]*Message, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	return s.messageRepo.FindConversationHistory(ctx, userID1, userID2, limit, beforeID)
}

// GetRecentConversations retrieves the latest message from each of the user's conversations.
func (s *messageService) GetRecentConversations(ctx context.Context, userID int64) ([]*Message, error) {
	return s.messageRepo.GetRecentConversations(ctx, userID)
}

// GetPendingMessages retrieves all messages for a user marked as 'PENDING'.
func (s *messageService) GetPendingMessages(ctx context.Context, userID int64) ([]*Message, error) {
	return s.messageRepo.FindPendingForUser(ctx, userID)
}

// MarkMessagesAsDelivered updates the status of a list of messages to 'DELIVERED'.
func (s *messageService) MarkMessagesAsDelivered(ctx context.Context, messageIDs []int64) error {
	if len(messageIDs) == 0 {
		return nil
	}
	return s.messageRepo.UpdateStatus(ctx, messageIDs, MessageDelivered)
}

// SendGroupMessage saves a message to the database and broadcasts it to all group members.
// FIX: Updated signature to accept mediaURL
func (s *messageService) SendGroupMessage(ctx context.Context, senderID int64, groupID int64, content string, mediaURL string) (*Message, error) {
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

	// 2. Determine Message Type
	messageType := TextMessage
	if mediaURL != "" {
		messageType = ImageMessage
	}
    // IMPORTANT: Assuming the message is either text OR media, not both requiring different types.

	// 3. Create the message struct
	message := &Message{
		SenderID:   senderID,
		RecipientID: groupID, // Recipient is the Group ID
		Type:        messageType,
		Content:     content,
                MediaURL:   mediaURL, 
		Timestamp:   time.Now(),
	}
    
    // 4. Input Validation (Safety Check)
    if message.Content == "" && message.MediaURL == "" {
        return nil, errors.New("message cannot be empty (no content or media URL provided)")
    }


	// 5. Save the message
	savedMessage, err := s.messageRepo.Save(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// 6. Broadcast the message to all group members (WebSocket Hub)
	s.hub.BroadcastGroupMessage(groupID, savedMessage)

	return savedMessage, nil
}


// SendP2PMessage saves a message to the database and broadcasts it to the recipient and sender.
func (s *messageService) SendP2PMessage(ctx context.Context, senderID int64, recipientID int64, content string, mediaURL string) (*Message, error) {
	// 1. Check if recipient user exists
	_, err := s.userRepo.GetByID(ctx, recipientID)
	if err != nil {
		if errors.Is(err, errors.New("user not found")) { // Assuming FindByID returns a "user not found" error
			return nil, errors.New("recipient user not found")
		}
		return nil, fmt.Errorf("failed to check recipient existence: %w", err)
	}

	// 2. Determine Message Type
	messageType := TextMessage
	if mediaURL != "" {
		messageType = ImageMessage
	}

	// 3. Create the message struct
	message := &Message{
		SenderID:    senderID,
		RecipientID: recipientID, // Recipient is the User ID
		Type:        messageType,
		Content:     content,
		MediaURL:    mediaURL,
		Timestamp:   time.Now(),
	}

	// 4. Input Validation (Safety Check)
	if message.Content == "" && message.MediaURL == "" {
		return nil, errors.New("message cannot be empty (no content or media URL provided)")
	}

	// 5. Save the message
	savedMessage, err := s.messageRepo.Save(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// 6. Broadcast the message to the sender and recipient (WebSocket Hub)
	// NOTE: We need to define BroadcastP2PMessage in the Hub interface next!
	s.hub.BroadcastP2PMessage(senderID, recipientID, savedMessage)

	return savedMessage, nil
}
