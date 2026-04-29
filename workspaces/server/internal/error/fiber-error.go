package apperror

import (
	"errors"

	"github.com/NishLy/go-fiber-boilerplate/internal/response"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

// httpStatus maps app error codes to HTTP status codes.
var httpStatus = map[Code]int{
	NotFound:         fiber.StatusNotFound,
	Duplicate:        fiber.StatusConflict,
	Invalid:          fiber.StatusBadRequest,
	Internal:         fiber.StatusInternalServerError,
	PermissionDenied: fiber.StatusForbidden,
	Unauthorized:     fiber.StatusUnauthorized,
}

func ErrorHandler(ctx fiber.Ctx, err error) error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return respondError(ctx, appErr)
	}

	// Fiber's own errors (e.g. 404 from unmatched routes, body parser failures)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return jsonError(ctx, fiberErr.Code, fiberErr.Message, nil)
	}

	// Truly unexpected — log with full detail
	logger.Sugar.Errorf("unhandled error: %+v", err)
	return jsonError(ctx, fiber.StatusInternalServerError, "internal server error", nil)
}

func respondError(ctx fiber.Ctx, e *Error) error {
	status, ok := httpStatus[e.Code]
	if !ok {
		status = fiber.StatusInternalServerError
	}

	// Only expose the message for client errors; mask internals
	msg := e.Message
	if status == fiber.StatusInternalServerError {
		logger.Sugar.Errorf("%v", e)
		msg = "internal server error"
	}

	return jsonError(ctx, status, msg, e.Data)
}

// jsonError is the single place that writes error responses.
func jsonError(ctx fiber.Ctx, status int, msg string, data interface{}) error {
	return ctx.Status(status).JSON(response.ErrorResponse{
		Code:  status,
		Error: msg,
		Data:  data,
	})
}
