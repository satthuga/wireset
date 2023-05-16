package middleware

import "github.com/gofiber/fiber/v2"

func AllowEmbed() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "*")
		c.Set("Expose-Headers", "X-Frame-Options,X-Shopify-API-Request-Failure-Reauthorize,X-Shopify-API-Request-Failure-Reauthorize")
		return c.Next()
	}
}
