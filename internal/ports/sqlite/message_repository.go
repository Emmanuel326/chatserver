package sqlite

import (
	"context"
	"log"
	"time" // Needed for time.Now() if you use it, though usually done in domain

	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/jmoiron/sqlx"
)

// messageRepository implements the domain.MessageRepository interface.
type messageRepository struct {
	db *sqlx.DB
}

// NewMessageRepository creates a new MessageRepository instance.
func NewMessageRepository(db *sqlx.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}

// Save persists a new message to the database. (Renamed from Create)
func (r *messageRepository) Save(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	// Ensure timestamp is set if the domain service didn't set it (safety check)
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	query := `
		INSERT INTO messages (sender_id, recipient_id, type, content, timestamp)
		VALUES (:sender_id, :recipient_id, :type, :content, :timestamp);
	`
	res, err := r.db.NamedExecContext(ctx, query, message)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	message.ID = id
	return message, nil
}

// FindConversationHistory retrieves the message history between two users. (Renamed from FindConversation)
func (r *messageRepository) FindConversationHistory(ctx context.Context, userID1, userID2 int64, limit int) ([]*domain.Message, error) {
	// Query to find messages where (sender=1 AND recipient=2) OR (sender=2 AND recipient=1)
	query := `
		SELECT id, sender_id, recipient_id, type, content, timestamp FROM messages
		WHERE (sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)
		ORDER BY timestamp DESC
		LIMIT ?;
	`
	messages := []*domain.Message{}
	// The arguments are passed to the query placeholders in order: userID1, userID2, userID2, userID1, limit
	err := r.db.SelectContext(ctx, &messages, query, userID1, userID2, userID2, userID1, limit)
	if err != nil {
		log.Printf("Error finding conversation history: %v", err)
		return nil, err
	}
	
	// Messages were fetched in DESC order, reverse them for chronological display
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
