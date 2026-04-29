package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func LimiterConfig() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        20,
		Expiration: 15 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).
				JSON(fiber.Map{
					"code":    fiber.StatusTooManyRequests,
					"status":  "error",
					"message": "Too many requests, please try again later",
				})
		},
		SkipSuccessfulRequests: true,
	})
}
