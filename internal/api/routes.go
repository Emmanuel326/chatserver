package api

import (
	"net/http"

	"github.com/Emmanuel326/chatserver/internal/api/middleware"
	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/Emmanuel326/chatserver/internal/ws"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all the API endpoints and injects dependencies.
func RegisterRoutes(router *gin.Engine, userService domain.UserService, jwtManager *auth.JWTManager, hub *ws.Hub, messageService domain.MessageService) {
	
	// Initialize Handlers (Dependency Injection)
<<<<<<< HEAD
	userHandler := NewUserHandler(userService, jwtManager)
	// FIX 1: Pass the jwtManager to the wsHandler constructor
	wsHandler := NewWSHandler(hub, jwtManager) 
	messageHandler := NewMessageHandler(messageService)
=======
	userHandler := NewUserHandler(userService, jwtManager) 
	wsHandler := NewWSHandler(hub) 
>>>>>>> 6ee29e9 (docs: removed verbose comments)

	// V1 API Group
	v1 := router.Group("/v1")
	{
		// --- Public User/Auth Routes ---
		v1.POST("/users/register", userHandler.Register)
		v1.POST("/users/login", userHandler.Login)

		// FIX 2: Move the WebSocket endpoint OUTSIDE the secured group
        // It authenticates via the 'token' query parameter inside the handler.
		v1.GET("/ws", wsHandler.ServeWs) 
        
		// --- Protected Routes Group ---
		// All routes inside this group will pass through the AuthMiddleware
		secured := v1.Group("/")
		secured.Use(middleware.AuthMiddleware(jwtManager))
		{
            // User Listing Endpoint (Secured)
            secured.GET("/users", userHandler.ListUsers) // Ensuring this is still here

			// Test Protected Endpoint
			secured.GET("/test-auth", func(c *gin.Context) {
				userID, _ := middleware.GetUserIDFromContext(c)
				c.JSON(http.StatusOK, gin.H{
					"message": "Access granted",
					"user_id": userID,
				})
			})
<<<<<<< HEAD

			//Message History endpoint
			secured.GET("/messages/history/:recipientID", messageHandler.GetConversationHistory)
=======
			
			// WebSocket Upgrade Endpoint
			secured.GET("/ws", wsHandler.ServeWs)
>>>>>>> 6ee29e9 (docs: removed verbose comments)
		}
	}
}
