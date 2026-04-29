package middleware

import (
	"context"
	"strings"

	"github.com/NishLy/go-fiber-boilerplate/config"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	t "github.com/NishLy/go-fiber-boilerplate/internal/token"
	pkg "github.com/NishLy/go-fiber-boilerplate/pkg/jwt"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

func Protected() fiber.Handler {
	cfg, err := config.Load()

	if err != nil {
		panic(err)
	}

	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		if token == "" {
			return apperror.UnauthorizedErr(nil, "Please authenticate")
		}

		userID, err := pkg.VerifyToken(token, cfg.JWT_SECRET, t.TokenTypeAccess)

		if err != nil {
			logger.Sugar.Debugf("Error verifying token: %v", err)
			return apperror.UnauthorizedErr(nil, "Please authenticate")
		}

		c.Locals("user_id", userID)
		// set user context for downstream handlers
		ctx := context.WithValue(c.Context(), "user_id", userID)
		c.SetContext(ctx)

		return c.Next()
	}
}
