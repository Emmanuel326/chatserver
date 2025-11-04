package domain

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

// UserService defines the methods for business logic related to users.
type UserService interface {
	// Methods implemented below:
	Register(ctx context.Context, username, email, password string) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ListAll(ctx context.Context) ([]*User, error)
	
	// FIXES REQUIRED BY main.go (for creating default users and checking):
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	CreateUser(ctx context.Context, username, password string) (*User, error) 
	
	// Other methods often required by other layers:
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	// NOTE: You may need to add or adjust other methods later
}

type userService struct {
	userRepo UserRepository
}

// NewUserService creates a new UserService instance.
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
func (s *userService) ListAll(ctx context.Context) ([]*User, error) {
	return s.userRepo.GetAll(ctx)
}

// --- NEW IMPLEMENTATIONS REQUIRED BY main.go ---

// GetUserByUsername implements the domain.UserService interface (for default user checking).
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	// Delegate to the Repository (assuming userRepo has a GetByUsername method)
	return s.userRepo.GetByUsername(ctx, username)
}

// CreateUser implements the domain.UserService interface (for default user creation).
func (s *userService) CreateUser(ctx context.Context, username, password string) (*User, error) {
	if username == "" || password == "" {
		return nil, &ValidationError{Msg: "Username and password are required"}
	}
	
	// Check if user already exists
	if _, err := s.userRepo.GetByUsername(ctx, username); err == nil {
		return nil, &ConflictError{Msg: "User with this username already exists"}
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// For testing, we use a placeholder email (your Register method requires it)
	user := NewUser(username, username+"@temp.com", string(hashedPassword))
	
	return s.userRepo.Create(ctx, user)
}

// GetUserByID implements the domain.UserService interface (needed by the WS hub).
func (s *userService) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	// Delegate to the Repository (assuming userRepo has a GetByID method)
	return s.userRepo.GetByID(ctx, userID)
}

// ListAllUsersWithChatInfo retrieves all users along with their last P2P message
// content and timestamp with the currentUserID.
func (s *userService) ListAllUsersWithChatInfo(ctx context.Context, currentUserID int64) ([]*UserWithChatInfo, error) {
	return s.userRepo.GetAllUsersWithLastMessageInfo(ctx, currentUserID)
}
