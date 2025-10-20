
// ------------------------------------------------------
// This file handles WebSocket connections for the chat.
// It upgrades HTTP requests to WebSocket connections,
// registers each connected client, and starts the read
// and write pumps for message exchange.
// ------------------------------------------------------

package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/Emmanuel326/chatserver/internal/ws"
)

// Create a shared Hub instance (global for all connections)
var hub = ws.NewHub()

// init starts the Hub in a goroutine when the package loads.
// This ensures the broadcast loop runs continuously.
func init() {
	go hub.Run()
}

// HandleWebSocket upgrades the HTTP request to a WebSocket connection.
func HandleWebSocket(c *fiber.Ctx) error {
	// Check if the incoming request is a WebSocket upgrade.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// ChatHandler manages each individual WebSocket connection.
// It will be passed to Fiber's websocket handler.
func ChatHandler(c *websocket.Conn) {
	client := &ws.Client{
		Conn: c,
		Send: make(chan []byte, 256),
	}

	hub.Register <- client
	log.Println("✅ Client connected:", c.RemoteAddr())

	// Start concurrent readers and writers for this client
	go client.WritePump()
	client.ReadPump(hub)

	log.Println("❌ Client disconnected:", c.RemoteAddr())
}

