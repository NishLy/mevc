package room

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

func RoomRouter(v1 fiber.Router, roomService RoomService) {
	roomHandler := NewRoomHandler(roomService)
	room := v1.Group("/rooms")

	room.Get("/:id", middleware.Protected(), roomHandler.GetRoomByID)
	room.Get("/code/:code", roomHandler.GetRoomByCode)
	room.Post("/upsert/:id?", middleware.Protected(), roomHandler.Upsert)
	room.Delete("/:id", middleware.Protected(), roomHandler.Delete)
}
