// main.go
// Entry point for the ChatServer project.
// This file simply wires up and launches the HTTP + WebSocket server.
// All logic lives inside internal packages (modular architecture).

package main

import (
	"log"

	// Import the server package — this sets up our Fiber app, routes, and middleware.
	"github.com/Emmanuel326/chatserver/internal/server"
)

func main() {
	// Initialize and configure the Fiber app using our custom server builder.
	app := server.New()

	// Start the server on port 8080.
	// Any fatal error (like port already in use) will be logged and stop execution.
	log.Fatal(app.Listen(":8080"))
}

