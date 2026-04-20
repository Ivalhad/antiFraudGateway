package middleware

import "github.com/gofiber/fiber/v2"

func APIKeyAuth(validKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-API-Key")

		if key == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Akses ditolak. Header X-API-Key tidak ditemukan.",
			})
		}

		if key != validKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Akses ditolak. API Key tidak valid.",
			})
		}

		return c.Next()
	}
}
