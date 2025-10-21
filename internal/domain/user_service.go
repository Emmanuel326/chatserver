package domain

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

// userService is the concrete implementation of the domain.UserService interface.
type userService struct {
	userRepo UserRepository
}

// NewUserService creates a new UserService instance.
// This is the function that main.go is looking for.
func NewUserService(repo UserRepository) UserService {
	return &userService{userRepo: repo}
}

// Register implements the domain.UserService interface.
func (s *userService) Register(ctx context.Context, username, email, password string) (*User, error) {
	// Simple validation
	if username == "" || email == "" || password == "" {
		return nil, &ValidationError{Msg: "Username, email, and password are required"}
	}
	
	// Check if user already exists
	if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		return nil, &ConflictError{Msg: "User with this email already exists"}
	}
	
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := NewUser(username, email, string(hashedPassword))
	
	// Delegate creation to the Repository
	return s.userRepo.Create(ctx, user)
}

// Authenticate implements the domain.UserService interface.
func (s *userService) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, &NotFoundError{Msg: "Invalid email or password"}
	}

	// Compare the stored hashed password with the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, &NotFoundError{Msg: "Invalid email or password"}
	}

	return user, nil
}

// ListAll retrieves all registered users.
// This implements the newly added method to the domain.UserService interface.
func (s *userService) ListAll(ctx context.Context) ([]*User, error) {
	return s.userRepo.GetAll(ctx)
}

// NOTE: You will also need NewUser() function and custom error types
// (like ValidationError, ConflictError, NotFoundError) which are typically defined in user.go
