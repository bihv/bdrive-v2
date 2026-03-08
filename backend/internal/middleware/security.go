package middleware

import "github.com/gofiber/fiber/v2"

// SetupSecurityHeaders adds security headers to all responses.
func SetupSecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "0")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		return c.Next()
	}
}
