
// ---------------------------------------------
// Represents the structure of messages sent/received
// over WebSocket (JSON friendly for any frontend).
// ---------------------------------------------

package ws

// Message defines the JSON structure for chat messages.
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
	Type     string `json:"type"` // e.g. "text", "image"
}

