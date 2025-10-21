package api

// RegisterRequest defines the expected JSON payload for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest defines the expected JSON payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse defines the standard response for successful authentication
type AuthResponse struct {
	Token string `json:"token"`
}
