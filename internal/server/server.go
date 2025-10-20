

// This package is responsible for initializing
// and configuring the Fiber app (HTTP server).
// It registers all HTTP and WebSocket routes and
// links them to the appropriate handlers.


package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	// Handlers contain the business logic for various endpoints.
	"github.com/Emmanuel326/chatserver/internal/handlers"
)

// New creates and returns a new Fiber app instance.
// Called from main.go — acts as the main entry point for
// both REST and WebSocket routes.
func New() *fiber.App {
	// Initialize a new Fiber app
	app := fiber.New()

	
	// ROOT / HEALTH ENDPOINT
	
	// Quick check that the server is live.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ChatServer is running 🚀")
	})

	// ----------------------
	// WEBSOCKET ENDPOINT
	// ------------------------
	// Step 1: HandleWebSocket checks if upgrade is allowed.
	// Step 2: websocket.New() actually upgrades the connection
	//         and delegates it to ChatHandler for per-client logic.
	app.Get("/ws", handlers.HandleWebSocket, websocket.New(handlers.ChatHandler))

	// ------------------------
	// FUTURE ENDPOINTS
	// -------------------
	// Example image uploads (POST):
	// app.Post("/upload", handlers.UploadHandler)
	//
	// Example REST API (GET messages):
	// app.Get("/api/messages", handlers.GetMessages)

	// Return the configured app so main.go can start it
	return app
}

