package api

// RegisterRequest defines the expected JSON payload for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"` // Clean whitespace before Email
	Password string `json:"password" binding:"required,min=8"` // Clean whitespace before Password
}

// LoginRequest defines the expected JSON payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"` // Clean whitespace before Email
	Password string `json:"password" binding:"required"` // Clean whitespace before Password
}

// AuthResponse defines the standard response for successful authentication
type AuthResponse struct {
	Token string `json:"token"`
}

// SendGroupMessageRequest defines the expected JSON payload for sending a group message.
// It includes the content for text messages and an optional MediaURL for images/files.
type SendGroupMessageRequest struct {
	Content string `json:"content"` // Text content (required for text message)
    MediaURL string `json:"media_url"` // Optional URL for external media (required for image/file message)
}
