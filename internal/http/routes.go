package http

import (
	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/Emmanuel326/chatserver/internal/handlers"
	"github.com/Emmanuel326/chatserver/internal/users"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// JWTMiddleware validates the Authorization header and attaches userID to context.
func JWTMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing Authorization token"})
	}

	userID, err := auth.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
	}

	// Store userID in Fiber's context for later access (e.g., WebSocket or handlers)
	c.Locals("userID", userID)
	return c.Next()
}

// SetupRoutes defines all HTTP routes
func SetupRoutes(app *fiber.App) {
	// Public routes
	app.Post("/register", users.RegisterHandler)
	app.Post("/login", users.LoginHandler)

	// ✅ Auth check route (optional, good for testing JWT)
	app.Get("/me", JWTMiddleware, func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(string)
		return c.JSON(fiber.Map{
			"message": "token valid",
			"user_id": userID,
		})
	})

	// ✅ Protected WebSocket endpoint
	app.Use("/ws", JWTMiddleware)
	app.Get("/ws", handlers.HandleWebSocket, websocket.New(handlers.ChatHandler))
}

