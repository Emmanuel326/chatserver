package users

// User represents a basic chat user.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"` // hashed, if you plan to secure
	Avatar   string `json:"avatar,omitempty"`
}

