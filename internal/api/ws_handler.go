package api

import (
	"log"
	"net/http"

	"github.com/Emmanuel326/chatserver/internal/api/middleware"
	"github.com/Emmanuel326/chatserver/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Upgrader defines the configuration for upgrading HTTP to WebSocket.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allows connections from any origin during development. MUST be restricted in production.
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

// WSHandler contains the dependency on the Hub.
type WSHandler struct {
	Hub *ws.Hub
}

// NewWSHandler creates a new handler instance.
func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{Hub: hub}
}

// ServeWs handles the HTTP request, upgrades the connection, and starts the WS client.
func (h *WSHandler) ServeWs(c *gin.Context) {
	// 1. Retrieve UserID from the Gin context (Set by the JWT Middleware)
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		// This should theoretically not happen if the route is protected by AuthMiddleware
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// 2. Upgrade HTTP Connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection for UserID %d: %v", userID, err)
		return
	}
	
	log.Printf("User %d successfully connected via WebSocket.", userID)

	// 3. Hand off the connection to the WS Client/Hub
	ws.ServeWs(h.Hub, conn, userID)
}
