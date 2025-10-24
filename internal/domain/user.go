package domain

import (
	"context"
	"time"
)

// The 'db' tag is for sqlx to map the struct fields to database columns.
type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // '-' means ignore in JSON output
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}


// NewUser is a constructor for creating a new User instance.
func NewUser(username, email, hashedPassword string) *User {
	return &User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}
}



// ValidationError is returned for invalid user input (e.g., missing fields).
type ValidationError struct {
	Msg string
}
func (e *ValidationError) Error() string { return "Validation Error: " + e.Msg }


type ConflictError struct {
	Msg string
}
func (e *ConflictError) Error() string { return "Conflict Error: " + e.Msg }

type NotFoundError struct {
	Msg string
}
func (e *NotFoundError) Error() string { return "Not Found Error: " + e.Msg }


// UserRepository defines the data access operations for users.
// This interface is implemented by the 'ports/sqlite' package.
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	// FIX: Add GetByUsername which is required by the UserService implementation
	GetByUsername(ctx context.Context, username string) (*User, error) 
	Create(ctx context.Context, user *User) (*User, error)
	GetAll(ctx context.Context) ([]*User, error) 
}

