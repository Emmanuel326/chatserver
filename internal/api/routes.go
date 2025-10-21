package api

import (
	"net/http"

	"github.com/Emmanuel326/chatserver/internal/api/middleware" // <--- NEW IMPORT
	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/Emmanuel326/chatserver/internal/domain"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all the API endpoints and injects dependencies.
func RegisterRoutes(router *gin.Engine, userService domain.UserService, jwtManager *auth.JWTManager) {
	
	// Initialize Handlers (Dependency Injection)
	userHandler := NewUserHandler(userService, jwtManager) 

	// V1 API Group
	v1 := router.Group("/v1")
	{
		// --- Public User/Auth Routes ---
		v1.POST("/users/register", userHandler.Register)
		v1.POST("/users/login", userHandler.Login)

		// --- Protected Routes Group ---
		// All routes inside this group will pass through the AuthMiddleware
		secured := v1.Group("/")
		secured.Use(middleware.AuthMiddleware(jwtManager)) // <--- Apply the JWT Middleware
		{
			// Test Protected Endpoint
			secured.GET("/test-auth", func(c *gin.Context) {
				userID, _ := middleware.GetUserIDFromContext(c)
				c.JSON(http.StatusOK, gin.H{
					"message": "Access granted",
					"user_id": userID, // Show the ID extracted from the token
				})
			})
			
			// TODO: protected.GET("/messages/history", messageHandler.GetHistory)
			// TODO: protected.GET("/ws", wsHandler.HandleWebSocket) // The final destination for the WS upgrade
		}
	}
}
