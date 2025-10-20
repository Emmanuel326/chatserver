package users

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// In-memory user store
var (
	usersMu sync.Mutex
	usersDB = make(map[string]*User)
)

// Register a new user (with bcrypt password hashing)
func Register(username, password string) (*User, error) {
	usersMu.Lock()
	defer usersMu.Unlock()

	// Check if username already exists
	for _, u := range usersDB {
		if u.Username == username {
			return nil, errors.New("username already exists")
		}
	}

	// 🔐 Hash the password before storing
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &User{
		ID:       uuid.NewString(),
		Username: username,
		Password: string(hashed),
	}
	usersDB[user.ID] = user
	return user, nil
}

// Authenticate verifies username/password using bcrypt
func Authenticate(username, password string) (*User, error) {
	usersMu.Lock()
	defer usersMu.Unlock()

	for _, u := range usersDB {
		if u.Username == username {
			// Compare the stored hash with the provided password
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
			if err == nil {
				return u, nil
			}
			break
		}
	}
	return nil, errors.New("invalid credentials")
}

