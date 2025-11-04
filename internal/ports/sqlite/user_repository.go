package sqlite

import (
	"context"
	"database/sql" // Needed for sql.ErrNoRows, sql.NullString, sql.NullTime, sql.NullInt64
	"log"

	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/jmoiron/sqlx"
)

// UserRepository implements the domain.UserRepository interface.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new repository instance.
func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := "SELECT id, username, email, password, created_at FROM users WHERE email = ?"
	
	err := r.db.GetContext(ctx, user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.NotFoundError{} // Return domain error for consistent service layer handling
		}
		return nil, err
	}
	return user, nil
}

// GetByID retrieves a user by their unique ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user := &domain.User{}
	query := "SELECT id, username, email, password, created_at FROM users WHERE id = ?"
	
	err := r.db.GetContext(ctx, user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.NotFoundError{}
		}
		return nil, err
	}
	return user, nil
}

// GetByUsername retrieves a user by their username. (NEW IMPLEMENTATION)
// This is required by the updated UserRepository interface in domain/user.go
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user := &domain.User{}
	query := "SELECT id, username, email, password, created_at FROM users WHERE username = ?"
	
	err := r.db.GetContext(ctx, user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.NotFoundError{}
		}
		return nil, err
	}
	return user, nil
}

// Create inserts a new user into the database and returns the created user (with ID).
func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := "INSERT INTO users (username, email, password, created_at) VALUES (?, ?, ?, ?)"
	
	result, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Password, user.CreatedAt)
	if err != nil {
		return nil, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user.ID = lastID
	return user, nil
}

// GetAll retrieves a list of all registered users.
func (r *UserRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	query := `
		SELECT id, username, email, created_at
		FROM users
		ORDER BY username ASC;
	`
	users := []*domain.User{}
	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return users, nil
		}
		log.Printf("Error retrieving all users: %v", err)
		return nil, err
	}
	return users, nil
}

// GetAllUsersWithLastMessageInfo retrieves all users, and for each user,
// includes the latest P2P message content and timestamp between that user
// and the currentUserID.
func (r *UserRepository) GetAllUsersWithLastMessageInfo(ctx context.Context, currentUserID int64) ([]*domain.UserWithChatInfo, error) {
	query := `
	WITH LastP2PMessages AS (
		SELECT
			m.id,
			m.sender_id,
			m.recipient_id,
			m.content,
			m.timestamp,
			CASE
				WHEN m.sender_id = :current_user_id THEN m.recipient_id
				ELSE m.sender_id
			END AS other_participant_id,
			ROW_NUMBER() OVER (
				PARTITION BY
					CASE
						WHEN m.sender_id = :current_user_id THEN m.recipient_id
						ELSE m.sender_id
					END
				ORDER BY m.timestamp DESC, m.id DESC
			) as rn
		FROM messages m
		WHERE
			m.type IN ('text', 'image') AND -- Only consider actual messages
			(m.sender_id = :current_user_id OR m.recipient_id = :current_user_id) -- Messages involving current user
	)
	SELECT
		u.id,
		u.username,
		u.email,
		u.created_at,
		lpm.content AS last_message_content,
		lpm.timestamp AS last_message_timestamp,
		lpm.sender_id AS last_message_sender_id
	FROM users u
	LEFT JOIN LastP2PMessages lpm ON u.id = lpm.other_participant_id AND lpm.rn = 1
	WHERE u.id != :current_user_id
	ORDER BY lpm.timestamp DESC, u.username ASC;
	`

	var usersWithInfo []*domain.UserWithChatInfo
	// Using Named Query for current_user_id
	rows, err := r.db.NamedQueryContext(ctx, query, map[string]interface{}{"current_user_id": currentUserID})
	if err != nil {
		log.Printf("Error retrieving all users with last message info for user %d: %v", currentUserID, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.UserWithChatInfo
		var lastMessageContent sql.NullString
		var lastMessageTimestamp sql.NullTime
		var lastMessageSenderID sql.NullInt64

		err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.CreatedAt,
			&lastMessageContent,
			&lastMessageTimestamp,
			&lastMessageSenderID,
		)
		if err != nil {
			log.Printf("Error scanning user with chat info row: %v", err)
			return nil, err
		}

		if lastMessageContent.Valid {
			u.LastMessageContent = &lastMessageContent.String
		}
		if lastMessageTimestamp.Valid {
			u.LastMessageTimestamp = &lastMessageTimestamp.Time
		}
		if lastMessageSenderID.Valid {
			u.LastMessageSenderID = &lastMessageSenderID.Int64
		}
		usersWithInfo = append(usersWithInfo, &u)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over user with chat info rows: %v", err)
		return nil, err
	}

	return usersWithInfo, nil
}
