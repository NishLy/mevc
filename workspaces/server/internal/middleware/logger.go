package middleware

import (
	"errors"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// 1. Execute the next handler
		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			status = fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				status = e.Code
			}
		}

		reqID := c.GetRespHeader("X-Request-ID")

		logger.Sugar.Info("http_request",
			zap.String("request_id", reqID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status), // Use the logic-derived status
			zap.Duration("latency", time.Since(start)),
		)

		return err
	}
}
