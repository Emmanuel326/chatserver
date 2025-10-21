package domain

import (
	"context"
	"time"
)

// User is the core data model for an application user.
// The 'db' tag is for sqlx to map the struct fields to database columns.
type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // '-' means ignore in JSON output
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}


// NewUser is a constructor for creating a new User instance.
// This resolves the 'undefined: NewUser' error in user_service.go.
func NewUser(username, email, hashedPassword string) *User {
	return &User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}
}


// Custom Error Types (Resolves all remaining 'undefined: XXXXError' errors)


// ValidationError is returned for invalid user input (e.g., missing fields).
type ValidationError struct {
	Msg string
}
func (e *ValidationError) Error() string { return "Validation Error: " + e.Msg }

// ConflictError is returned when a resource already exists (e.g., email taken).
type ConflictError struct {
	Msg string
}
func (e *ConflictError) Error() string { return "Conflict Error: " + e.Msg }

// NotFoundError is returned when a user or resource cannot be found.
type NotFoundError struct {
	Msg string
}
func (e *NotFoundError) Error() string { return "Not Found Error: " + e.Msg }



// UserService defines the business operations related to users.
// This interface is implemented by the concrete 'user' services.
type UserService interface {
	Register(ctx context.Context, username, email, password string) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ListAll(ctx context.Context) ([]*User, error)
}

// UserRepository defines the data access operations for users.
// This interface is implemented by the 'ports/sqlite' package.
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	GetAll(ctx context.Context) ([]*User, error) // <-- FIX: space added between ctx and context.Context
}
