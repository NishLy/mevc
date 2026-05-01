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
		path := c.Path()
		if path == "/ws" || path == "/ws/" || path == "/" {
			return c.Next()
		}

		method := c.Method()
		logger.Sugar.Infof("Incoming request: %s %s", method, path)
		start := time.Now()
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
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
		)

		return err
	}
}
