package utils

import "github.com/gofiber/fiber/v3"

func Success(c fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func Error(c fiber.Ctx, err string) error {
	return c.Status(500).JSON(fiber.Map{
		"success": false,
		"error":   err,
	})
}

func NotFound(c fiber.Ctx, message string) error {
	return c.Status(404).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

func BadRequest(c fiber.Ctx, message string) error {
	return c.Status(400).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

func Unauthorized(c fiber.Ctx, message string) error {
	return c.Status(401).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

func Forbidden(c fiber.Ctx, message string) error {
	return c.Status(403).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

func InternalServerError(c fiber.Ctx, message string) error {
	return c.Status(500).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}
