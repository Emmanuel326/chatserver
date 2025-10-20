package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Emmanuel326/chatserver/internal/handlers"
)

func New() *fiber.App {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx)*

