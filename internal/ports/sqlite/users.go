package sqlite

import (
	"context"

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
	
	// sqlx.Get is a convenience function that executes the query and loads the result into the struct.
	err := r.db.GetContext(ctx, user, query, email)
	
	// A common pattern: if err == sql.ErrNoRows, you'd return a specific domain error
	// (for now, we'll return the raw error for simplicity).
	if err != nil {
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
		return nil, err
	}
	return user, nil
}

// Create inserts a new user into the database and returns the created user (with ID).
func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := "INSERT INTO users (username, email, password, created_at) VALUES (?, ?, ?, ?)"
	
	// For SQLite, the last inserted ID can be retrieved by executing the query directly.
	result, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Password, user.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Get the last inserted ID
	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update the user struct with the new ID
	user.ID = lastID

	return user, nil
}
