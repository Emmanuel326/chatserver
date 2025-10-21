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
func RegisterRoutes(router *gin.Engine, userService domain.UserService, jwtManager *auth.JWTManager, hub *ws.Hub) {
	
	// Initialize Handlers (Dependency Injection)
	userHandler := NewUserHandler(userService, jwtManager) 
	wsHandler := NewWSHandler(hub) // <--- FIX: wsHandler is now initialized here

	// V1 API Group
	v1 := router.Group("/v1")
	{
		// --- Public User/Auth Routes ---
		v1.POST("/users/register", userHandler.Register)
		v1.POST("/users/login", userHandler.Login)

		// --- Protected Routes Group ---
		// All routes inside this group will pass through the AuthMiddleware
		secured := v1.Group("/")
		secured.Use(middleware.AuthMiddleware(jwtManager)) 
		{
			// Test Protected Endpoint
			secured.GET("/test-auth", func(c *gin.Context) {
				userID, _ := middleware.GetUserIDFromContext(c)
				c.JSON(http.StatusOK, gin.H{
					"message": "Access granted",
					"user_id": userID, 
				})
			})
			
			// WebSocket Upgrade Endpoint
			secured.GET("/ws", wsHandler.ServeWs) // <--- wsHandler is now defined
		}
	}
}

