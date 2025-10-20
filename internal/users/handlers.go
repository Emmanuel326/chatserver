package users

import (
	"github.com/Emmanuel326/chatserver/internal/auth"
	"github.com/gofiber/fiber/v2"
)

// RegisterHandler handles POST /register
func RegisterHandler(c *fiber.Ctx) error {
	req := new(User)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := Register(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "registration successful",
		"user":    user,
	})
}

// LoginHandler handles POST /login
func LoginHandler(c *fiber.Ctx) error {
	req := new(User)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := Authenticate(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// ✅ Generate JWT instead of mock token
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"message": "login successful",
		"token":   token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

