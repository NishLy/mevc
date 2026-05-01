package middleware

import (
	"context"

	"github.com/NishLy/go-fiber-boilerplate/internal/platform/database"
	"github.com/gofiber/fiber/v3"
)

// InjectTenantIdentifier is a middleware that injects the tenant identifier into the request context. It expects the tenant identifier to be provided in the "X-Tenant-ID" header of the request.
func InjectTenantIdentifier() fiber.Handler {
	return func(c fiber.Ctx) error {
		tenantID := c.Get("X-Tenant-ID")

		// whitelist /swagger/* and /docs/* for testing purposes
		path := c.Path()
		if path == "/swagger" || path == "/docs" || (len(path) > 8 && path[:8] == "/swagger/") || (len(path) > 6 && path[:6] == "/docs/") {
			return c.Next()
		}

		if path == "/ws" {
			return c.Next()
		}

		if tenantID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "X-Tenant-ID header is required",
			})
		}

		db, err := database.GetDB(tenantID, true)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to connect to database",
			})
		}

		ctx := context.WithValue(c.Context(), "db", db.DB)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		c.SetContext(ctx)
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}
}

// fiber:context-methods migrated
