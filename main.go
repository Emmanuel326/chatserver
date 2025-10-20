package main

import (
	"log"

	"github.com/Emmanuel326/chatserver/internal/server"
)

func main() {
	app := server.New() // builds the Fiber app with routes
	log.Fatal(app.Listen(":8080"))
}

