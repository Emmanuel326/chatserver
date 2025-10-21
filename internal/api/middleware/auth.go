package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/gin-gonic/gin"
)

// Gin Context Key for storing UserID
const ContextUserIDKey = "userID"

// AuthMiddleware is a Gin middleware that validates the JWT from the Authorization header.
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract Token from Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Header format expected: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization format. Expected 'Bearer <token>'"})
			c.Abort()
			return
		}
		
		tokenString := parts[1]

		// 2. Validate Token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			log.Printf("JWT validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 3. Set UserID in Gin Context
		// The UserID is now available to all downstream handlers in this request chain
		c.Set(ContextUserIDKey, claims.UserID)

		// 4. Continue to the next handler/logic
		c.Next()
	}
}

// GetUserIDFromContext is a helper function for handlers to safely retrieve the UserID.
func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	idValue, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}
	
	userID, ok := idValue.(int64)
	if !ok {
		return 0, false
	}
	return userID, true
}
