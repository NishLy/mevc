package routes

import (
	swagger "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
)

func DocsRoutes(v1 fiber.Router) {
	docs := v1.Group("/docs")

	docs.Get("/*", swagger.HandlerDefault)
}
