package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/config"
	"github.com/NishLy/go-fiber-boilerplate/internal/app"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"github.com/NishLy/go-fiber-boilerplate/internal/middleware"
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/database"
	"github.com/NishLy/go-fiber-boilerplate/internal/room"
	"github.com/NishLy/go-fiber-boilerplate/internal/routes"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"go.uber.org/zap"
)

func main() {
	logger.Init()

	configApp, err := config.Load()

	if err != nil {
		logger.Sugar.Fatal("Failed to load config", zap.Error(err))
	}

	fiberApp := fiber.New(fiber.Config{ErrorHandler: apperror.ErrorHandler})
	io := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: func(r *http.Request) bool { return true },
			},
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool { return true },
			},
		},
	})

	room.RoomWsBootstrap(io)
	go func() {
		if err := io.Serve(); err != nil {
			logger.Sugar.Fatal("Socket.IO server error:", zap.Error(err))
		}
	}()
	defer io.Close()

	// Use adaptor.HTTPHandler with the full mux, not just io directly
	mux := http.NewServeMux()
	mux.HandleFunc("/socket.io/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		io.ServeHTTP(w, r)
	})

	// Run both servers concurrently
	go func() {
		logger.Sugar.Info("Starting Socket.IO server on :8001")
		if err := http.ListenAndServe(":8001", mux); err != nil {
			logger.Sugar.Fatal("Socket.IO server error:", zap.Error(err))
		}
	}()

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
