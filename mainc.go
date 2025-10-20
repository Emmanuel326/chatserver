package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()

	// Serve a simple test page for convenience
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ChatServer is running 🚀\nGo to /ws (WebSocket endpoint)")
	})

	// Only allow WebSocket upgrade if request header matches
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read error:", err)
				break
			}
			log.Printf("recv: %s", msg)
			if err = c.WriteMessage(mt, append([]byte("Server: "), msg...)); err != nil {
				log.Println("write error:", err)
				break
			}
		}
	}))

	log.Fatal(app.Listen(":8080"))
}

