package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/gin-gonic/gin"
)

// UserHandler contains the dependencies required by user API endpoints.
type UserHandler struct {
	UserService domain.UserService
	JWTManager  *auth.JWTManager
}

// NewUserHandler creates a new handler instance.
func NewUserHandler(userService domain.UserService, jwtManager *auth.JWTManager) *UserHandler {
	return &UserHandler{
		UserService: userService,
		JWTManager:  jwtManager,
	}
}

// ListUsersWithChatInfo handles GET /v1/users/with-chat-info to retrieve
// all registered users along with the last P2P message content and timestamp
// for the authenticated user with each listed user.
func (h *UserHandler) ListUsersWithChatInfo(c *gin.Context) {
	currentUserID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed or user ID missing"})
		return
	}

	usersWithInfo, err := h.UserService.ListAllUsersWithChatInfo(c.Request.Context(), currentUserID)
	if err != nil {
		log.Printf("Failed to retrieve users with chat info for user %d: %v", currentUserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user chat list"})
		return
	}

	// Map domain.UserWithChatInfo to api.UserCardResponse
	response := make([]UserCardResponse, len(usersWithInfo))
	for i, u := range usersWithInfo {
		response[i] = UserCardResponse{
			ID:                   u.ID,
			Username:             u.Username,
			Email:                u.Email,
			LastMessageContent:   u.LastMessageContent,
			LastMessageTimestamp: u.LastMessageTimestamp,
			LastMessageSenderID:  u.LastMessageSenderID,
		}
	}

	c.JSON(http.StatusOK, response)
}

// Register handles user registration via HTTP POST.
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format or missing fields"})
		return
	}

	user, err := h.UserService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		log.Printf("Registration failed for %s: %v", req.Email, err)

		// ✅ Type-safe error handling (no brittle string compares)
		switch e := err.(type) {
		case *domain.ValidationError:
			c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
			return
		case *domain.ConflictError:
			c.JSON(http.StatusConflict, gin.H{"error": e.Error()})
			return
		default:
			// fallback if error message isn’t wrapped properly
			if strings.Contains(strings.ToLower(err.Error()), "exists") {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	token, err := h.JWTManager.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration successful, but failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
	})
}

// ListUsers handles GET /v1/users to retrieve all registered users.
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.UserService.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user list"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Login handles user authentication via HTTP POST.
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format or missing fields"})
		return
	}

	user, err := h.UserService.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("Authentication failed for %s: %v", req.Email, err)
		switch e := err.(type) {
		case *domain.ValidationError:
			c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
			return
		case *domain.NotFoundError:
			c.JSON(http.StatusUnauthorized, gin.H{"error": e.Error()})
			return
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
	}

	token, err := h.JWTManager.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login successful, but failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

