package sqlite

import (
	"context"
	"database/sql" // Needed for sql.ErrNoRows
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
// This implements the final missing method from the domain.UserRepository interface.
func (r *UserRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	// IMPORTANT: Exclude the password column from the selection
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
