
// This package initializes and configures the Fiber
// app (HTTP server). It registers all HTTP and
// WebSocket routes and links them to handlers.


package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/Emmanuel326/chatserver/internal/users"

	// Local imports
	"github.com/Emmanuel326/chatserver/internal/http"
	"github.com/Emmanuel326/chatserver/internal/handlers"
)

// New creates and returns a fully configured Fiber app.
// It is called from main.go — acts as the main entry point
// for both REST and WebSocket endpoints.
func New() *fiber.App {
	// Initialize Fiber with custom config (e.g. increased upload limit)
	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024, // 20 MB max request body (uploads)
		AppName:   "ChatServer 🚀",
	})

	// ----------------------
	// ROOT / HEALTH CHECK
	// ----------------------
	// Basic check to confirm the server is running.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ChatServer is running 🚀")
	})

	// ----------------------
	// WEBSOCKET ENDPOINT
	// ----------------------
	// Step 1: HandleWebSocket ensures upgrade conditions.
	// Step 2: websocket.New(...) upgrades to WS and manages per-client logic.
	app.Get("/ws", handlers.HandleWebSocket, websocket.New(handlers.ChatHandler))

	// ----------------------
	// FILE UPLOAD ENDPOINT
	// ----------------------
	// Handles POST /upload for image or file uploads.
	app.Post("/upload", http.UploadHandler)

	// ----------------------
	// STATIC FILES
	// ----------------------
	// Serves uploaded files from the local /uploads directory.
	app.Static("/uploads", "./uploads")
	app.Post("/register", users.RegisterHandler)
        app.Post("/login", users.LoginHandler)

	// ----------------------
	// (Optional future routes)
	// ----------------------
	// app.Get("/api/messages", handlers.GetMessages)
	// app.Post("/api/send", handlers.SendMessage)

	// Return the configured app instance to main.go
	return app
}

