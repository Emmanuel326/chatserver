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
func RegisterRoutes(
	router *gin.Engine,
	userService domain.UserService,
	jwtManager *auth.JWTManager,
	hub *ws.Hub,
	messageService domain.MessageService,
	groupService domain.GroupService,
) {
	
	// Initialize Handlers (Dependency Injection)
	// NOTE: We assume these constructors exist in internal/api/
	userHandler := NewUserHandler(userService, jwtManager)
	wsHandler := NewWSHandler(hub, jwtManager)
	messageHandler := NewMessageHandler(messageService)
	groupHandler := NewGroupHandler(groupService)

	// --- WebSocket Route ---
	// The client connects to /ws, so it must be outside the /v1 group
	router.GET("/ws", wsHandler.ServeWs)

	// V1 API Group
	v1 := router.Group("/v1")
	{
		// --- Public User/Auth Routes (Registration and Login) ---
		v1.POST("/users/register", userHandler.Register)
		// Confirmed login route is POST /v1/users/login
		v1.POST("/users/login", userHandler.Login)


		// --- Protected Routes Group ---
		secured := v1.Group("/")
		secured.Use(middleware.AuthMiddleware(jwtManager))
		{
 			// User Listing Endpoint (FIXED: Removed invalid character and cleaned indentation)
			secured.GET("/users", userHandler.ListUsers)

			// Message History endpoint
			secured.GET("/messages/history/:recipientID", messageHandler.GetConversationHistory)
			
			// Group Endpoints
			secured.POST("/groups", groupHandler.CreateGroup)
			secured.POST("/groups/:groupID/members", groupHandler.AddMember)
			// ADDED: Missing GetMembers route for completeness
			secured.GET("/groups/:groupID/members", groupHandler.GetMembers)

			// Message Send Endpoint (via API) - The target of our final test
			secured.POST("/messages/group/:groupID", messageHandler.SendGroupMessage)

			// Test Protected Endpoint
			secured.GET("/test-auth", func(c *gin.Context) {
				userID, _ := middleware.GetUserIDFromContext(c)
				c.JSON(http.StatusOK, gin.H{
					"message": "Access granted",
					"user_id": userID,
				})
			})
		}
	}
}
