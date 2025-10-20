package ws

// Message defines the structure of chat messages
type Message struct {
	Type      string `json:"type"`      // e.g. "text" or "image"
	Username  string `json:"username"`  // who sent it
	Content   string `json:"content"`   // message text or image URL
	Timestamp int64  `json:"timestamp"` // unix time for ordering
}

