package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"Message": "Hello, World!"})
}
