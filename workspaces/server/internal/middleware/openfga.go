package middleware

import (
	"context"

	fga "github.com/NishLy/go-fiber-boilerplate/internal/openfga"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

func InjectOpenFGA() fiber.Handler {
	return func(c fiber.Ctx) error {
		tenantID := c.Get("X-Tenant-ID")

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

		fgaClient, err := fga.GetFGAClient(tenantID)
		if err != nil {
			logger.Sugar.Errorf("Failed to get OpenFGA client for tenant %s: %v", tenantID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to connect to OpenFGA",
			})
		}

		ctx := context.WithValue(c.Context(), "fga", fgaClient)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		c.SetContext(ctx)
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}
}

// fiber:context-methods migrated
