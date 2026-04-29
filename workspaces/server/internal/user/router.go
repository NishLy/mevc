package user

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/middleware"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

func UserRouter(v1 fiber.Router, userService *UserService) {
	userHandler := NewUserHandler(logger.Sugar, *userService)
	user := v1.Group("/users", middleware.Protected())

	user.Get("/", userHandler.GetUsers)
}
