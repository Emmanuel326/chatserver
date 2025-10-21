package api

import (
	"log"
	"net/http"

	"github.com/Emmanuel326/chatserver/internal/auth" 
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/gin-gonic/gin"
)

// UserHandler contains the dependencies required by user API endpoints.
type UserHandler struct {
	UserService domain.UserService 
	JWTManager  *auth.JWTManager // <--- NEW: JWT dependency
}

// NewUserHandler creates a new handler instance.
// Now receives both UserService and JWTManager.
func NewUserHandler(userService domain.UserService, jwtManager *auth.JWTManager) *UserHandler {
	return &UserHandler{
		UserService: userService, 
		JWTManager:  jwtManager, // <--- Dependency injected
	}
}

// Register handles user registration via HTTP POST.
func (h *UserHandler) Register(c *gin.Context) {
	// ... (Binding and Validation remains the same)
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format or missing fields"})
		return
	}

	// 1. Call Domain Service (Business Logic)
	user, err := h.UserService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		log.Printf("Registration failed for %s: %v", req.Email, err)
		if err.Error() == "user already exists with this email" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 2. Generate JWT Token on successful registration
	token, err := h.JWTManager.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration successful, but failed to generate token"})
		return
	}
	
	// 3. Success Response with Token
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token": token, // <--- NOW RETURNS THE TOKEN
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
	// ... (Binding and Validation remains the same)
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format or missing fields"})
		return
	}

	// 1. Call Domain Service (Business Logic)
	user, err := h.UserService.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("Authentication failed for %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// 2. Generate JWT Token on successful login
	token, err := h.JWTManager.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login successful, but failed to generate token"})
		return
	}
	
	// 3. Success Response with Token
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token": token, // <--- NOW RETURNS THE TOKEN
	})
}
