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



// UserService defines the business operations related to users.
// This interface is implemented by the concrete 'user' services.
type UserService interface {
	Register(ctx context.Context, username, email, password string) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
}

// UserRepository defines the data access operations for users.
// This interface is implemented by the 'ports/sqlite' package.
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
}
