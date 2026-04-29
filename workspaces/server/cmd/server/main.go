package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/config"
	"github.com/NishLy/go-fiber-boilerplate/internal/app"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"github.com/NishLy/go-fiber-boilerplate/internal/middleware"
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/database"
	"github.com/NishLy/go-fiber-boilerplate/internal/routes"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	socketio "github.com/googollee/go-socket.io"
	"go.uber.org/zap"
)

func main() {
	logger.Init()

	configApp, err := config.Load()

	if err != nil {
		logger.Sugar.Fatal("Failed to load config", zap.Error(err))
	}

	fiberApp := fiber.New(fiber.Config{ErrorHandler: apperror.ErrorHandler})
	io := socketio.NewServer(nil)
	defer io.Close()

	// Register Socket.IO server as a handler for the "/socket.io/*" route
	fiberApp.All("/socket.io/*", adaptor.HTTPHandler(io))

	// Start a goroutine to periodically clean up idle database connections
	database.CleanupDBs(time.Second * 60)

	// middlewares
	fiberApp.Use(helmet.New())
	fiberApp.Use(compress.New())
	fiberApp.Use(middleware.Logger())
	fiberApp.Use(requestid.New())
	fiberApp.Use(middleware.RecoverConfig())
	fiberApp.Use(middleware.LimiterConfig())
	fiberApp.Use(middleware.CORSConfig())
	fiberApp.Use(middleware.InjectTenantIdentifier())
	fiberApp.Use(middleware.InjectOpenFGA())

	appContainer := &app.App{
		Config: configApp,
		Io:     io,
	}

	routes.Setup(appContainer, fiberApp)

	logger.Sugar.Info("Starting server on " + configApp.HOST + ":" + configApp.PORT)

	// Server configuration
	address := configApp.HOST + ":" + configApp.PORT
	// Channel to capture server errors
	serverErrors := make(chan error, 1)
	// Start server in a separate goroutine
	go startServer(fiberApp, address, serverErrors)
	// Handle graceful shutdown and server errors
	handleGracefulShutdown(context.Background(), fiberApp, serverErrors)
}

func startServer(app *fiber.App, address string, serverErrors chan<- error) {
	if err := app.Listen(address, fiber.ListenConfig{
		EnablePrefork: config.Get().ENV == "production",
	}); err != nil {
		serverErrors <- err
	}
}

func handleGracefulShutdown(ctx context.Context, app *fiber.App, serverErrors <-chan error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sugar := logger.Sugar

	select {
	case err := <-serverErrors:
		sugar.Fatalf("Server error: %v", err)
	case <-quit:
		sugar.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			sugar.Fatalf("Error during server shutdown: %v", err)
		}
	case <-ctx.Done():
		sugar.Info("Context cancelled, shutting down server...")
	}

	sugar.Info("Server gracefully stopped")
}
