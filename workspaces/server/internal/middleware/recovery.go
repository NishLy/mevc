package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func RecoverConfig() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
	})
}
