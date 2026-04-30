package routes

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/app"
	"github.com/NishLy/go-fiber-boilerplate/internal/auth"
	"github.com/NishLy/go-fiber-boilerplate/internal/token"
	"github.com/NishLy/go-fiber-boilerplate/internal/user"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/fiber/v3"
)

func Setup(appContainer *app.App, app *fiber.App) {
	api := app.Group("/api").Group("/v1")

	sugarLogger := logger.Sugar

	// init user repository and service
	userRepo := user.NewUserRepository(*sugarLogger)
	userService := user.NewUserService(userRepo, *sugarLogger)

	// init token service
	tokenRepo := token.NewTokenRepository(*sugarLogger)
	tokenService := token.NewTokenService(*sugarLogger, tokenRepo)

	// init  auth handler and service
	authService := auth.NewAuthService(userService, tokenService)
	auth.AuthRouter(api, authService)

	// init user handler and service
	user.UserRouter(api, &userService)

	// init docs route
	if appContainer.Config.ENV == "development" {
		sugarLogger.Info("Running in development mode, enabling Swagger docs")
		DocsRoutes(api)
	}

	// app.Get("/ws", websocket.New(ws.Handler(appContainer.WsHub)))

	// init room handler and service

}
