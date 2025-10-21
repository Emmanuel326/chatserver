package domain

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Concrete implementation of the domain.UserService interface
type userService struct {
	userRepo UserRepository
	// We'll add JWT generating dependency here later
}

// NewUserService creates a new UserService instance.
// It receives the UserRepository as a dependency (Dependency Injection).
func NewUserService(repo UserRepository) UserService {
	return &userService{userRepo: repo}
}

// Register creates a new user, hashes the password, and saves it.
func (s *userService) Register(ctx context.Context, username, email, password string) (*User, error) {
	// 1. Check for existing user by email (business rule)
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		// This is a custom domain error you'd define, but for JIT we use built-in error
		return nil, errors.New("user already exists with this email") 
	}
	// We should check if the error is sql.ErrNoRows, but for simplicity, we proceed

	// 2. Hash the password (Security rule from C/Rust background!)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 3. Create the User model
	user := &User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	// 4. Persist (delegates to the Ports layer)
	return s.userRepo.Create(ctx, user)
}

// Authenticate verifies credentials and returns the user.
func (s *userService) Authenticate(ctx context.Context, email, password string) (*User, error) {
	// 1. Retrieve the user by email (delegates to the Ports layer)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// A database-related error, potentially "no rows"
		return nil, errors.New("invalid credentials")
	}

	// 2. Compare the stored hash with the provided password (Security rule)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// bcrypt.CompareHashAndPassword returns an error if they don't match
		return nil, errors.New("invalid credentials")
	}

	// Success: return the authenticated user
	return user, nil
}
