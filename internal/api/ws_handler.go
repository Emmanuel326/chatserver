package api

import (
	"log"
	"net/http"
	"github.com/Emmanuel326/chatserver/internal/auth" 
	"github.com/Emmanuel326/chatserver/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)


// Define the global upgrader instance
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


// WSHandler contains the dependency on the Hub and JWT Manager.
type WSHandler struct {
	Hub *ws.Hub
	jwtManager *auth.JWTManager // <-- Must be used for validation
}

// NewWSHandler remains the same (Correct)
func NewWSHandler(hub *ws.Hub, jwtManager *auth.JWTManager) *WSHandler {
	return &WSHandler{
		Hub: hub,
		jwtManager: jwtManager,
	}
}

// ServeWs handles the HTTP request, performs authentication, upgrades the connection, and starts the WS client.
func (h *WSHandler) ServeWs(c *gin.Context) {
    // 1. Extract token from query parameter
    tokenString := c.Query("token")
    if tokenString == "" {
        log.Println("WS Connect attempt: Missing token query parameter")
        c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
        return
    }

    // 2. Validate token and get User ID
    claims, err := h.jwtManager.ValidateToken(tokenString) // Assuming your JWTManager exposes ValidateToken
    if err != nil {
        log.Printf("WS Connect attempt: Invalid token: %v", err)
        c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
        return
    }
    userID := claims.UserID // Ensure claims struct has UserID

	// 3. Upgrade HTTP Connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection for UserID %d: %v", userID, err)
		return
	}
	
	log.Printf("User %d successfully connected via WebSocket.", userID)

	
	ws.ServeWs(h.Hub, conn, userID) 
}
