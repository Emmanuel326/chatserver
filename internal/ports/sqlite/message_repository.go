package sqlite

import (
	"context"
	"log"
	"time"

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

// Save persists a new message to the database.
func (r *messageRepository) Save(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	query := `
		INSERT INTO messages (sender_id, recipient_id, type, content, media_url, timestamp, status)
		VALUES (:sender_id, :recipient_id, :type, :content, :media_url, :timestamp, :status);
	`
    // FIX: NamedExecContext automatically maps message.MediaURL to :media_url
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

// FindConversationHistory retrieves the message history between two users with pagination.
func (r *messageRepository) FindConversationHistory(ctx context.Context, userID1, userID2 int64, limit int, beforeID int64) ([]*domain.Message, error) {
	// Base query
	query := `
		SELECT id, sender_id, recipient_id, type, content, media_url, timestamp, status FROM messages
		WHERE ((sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?))
	`
	args := []interface{}{userID1, userID2, userID2, userID1}

	// Add pagination condition
	if beforeID > 0 {
		query += " AND id < ?"
		args = append(args, beforeID)
	}

	// Add ordering and limit
	query += `
		ORDER BY id DESC
		LIMIT ?;
	`
	args = append(args, limit)

	messages := []*domain.Message{}
	err := r.db.SelectContext(ctx, &messages, query, args...)
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

// GetGroupConversationHistory retrieves the message history for a group.
func (r *messageRepository) GetGroupConversationHistory(ctx context.Context, groupID int64, limit int) ([]*domain.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, type, content, media_url, timestamp, status FROM messages
		WHERE recipient_id = ?
		ORDER BY timestamp DESC
		LIMIT ?;
	`
	messages := []*domain.Message{}
	err := r.db.SelectContext(ctx, &messages, query, groupID, limit)
	if err != nil {
		log.Printf("Error finding group conversation history: %v", err)
		return nil, err
	}

	// Messages were fetched in DESC order, reverse them for chronological display
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetRecentConversations returns the latest message for every unique conversation
// a user has participated in (both P2P and Group chats).
func (r *messageRepository) GetRecentConversations(ctx context.Context, userID int64) ([]*domain.Message, error) {
	query := `
		-- CTE for groups the user is a member of
		WITH user_groups AS (
		  SELECT group_id FROM group_members WHERE user_id = ?
		),
		-- All relevant messages for the user, with a generated conversation ID
		messages_with_conv_id AS (
		  SELECT
			m.id, m.sender_id, m.recipient_id, m.type, m.content, m.media_url, m.timestamp, m.status,
			CASE
			  -- Group message where user is a member
			  WHEN g.id IS NOT NULL AND m.recipient_id IN (SELECT group_id FROM user_groups)
				THEN 'group_' || m.recipient_id
			  -- P2P message sent by user
			  WHEN g.id IS NULL AND m.sender_id = ?
				THEN 'user_' || m.recipient_id
			  -- P2P message received by user
			  WHEN g.id IS NULL AND m.recipient_id = ?
				THEN 'user_' || m.sender_id
			  ELSE NULL
			END AS conv_id
		  FROM messages m
		  LEFT JOIN groups g ON m.recipient_id = g.id
		  WHERE
			-- It's a group message and the user is a member
			(g.id IS NOT NULL AND m.recipient_id IN (SELECT group_id FROM user_groups))
			OR
			-- or it's a P2P message involving the user
			(g.id IS NULL AND (m.sender_id = ? OR m.recipient_id = ?))
		),
		-- Rank messages within each conversation by timestamp
		ranked_messages AS (
		  SELECT
			*,
			ROW_NUMBER() OVER(PARTITION BY conv_id ORDER BY timestamp DESC) as rn
		  FROM messages_with_conv_id
		  WHERE conv_id IS NOT NULL
		)
		-- Select only the latest message from each conversation
		SELECT id, sender_id, recipient_id, type, content, media_url, timestamp, status
		FROM ranked_messages
		WHERE rn = 1
		ORDER BY timestamp DESC;
	`

	messages := []*domain.Message{}
	err := r.db.SelectContext(ctx, &messages, query, userID, userID, userID, userID, userID)
	if err != nil {
		log.Printf("Error finding recent conversations: %v", err)
		return nil, err
	}

	return messages, nil
}

// FindPendingForUser retrieves all messages for a user with 'PENDING' status.
func (r *messageRepository) FindPendingForUser(ctx context.Context, userID int64) ([]*domain.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, type, content, media_url, timestamp, status
		FROM messages
		WHERE recipient_id = ? AND status = ?
		ORDER BY timestamp ASC;
	`
	messages := []*domain.Message{}
	err := r.db.SelectContext(ctx, &messages, query, userID, domain.MessagePending)
	if err != nil {
		log.Printf("Error finding pending messages for user %d: %v", userID, err)
		return nil, err
	}
	return messages, nil
}

// UpdateStatus performs a bulk update on the status of given message IDs.
func (r *messageRepository) UpdateStatus(ctx context.Context, messageIDs []int64, status domain.MessageStatus) error {
	if len(messageIDs) == 0 {
		return nil // Nothing to update
	}

	query, args, err := sqlx.In("UPDATE messages SET status = ? WHERE id IN (?);", status, messageIDs)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error updating message statuses: %v", err)
	}
	return err
}
