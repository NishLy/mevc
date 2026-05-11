package schedule

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

func ScheduleRouter(v1 fiber.Router, scheduleService ScheduleService) {
	scheduleHandler := NewScheduleHandler(scheduleService)
	schedule := v1.Group("/schedules", middleware.Protected())

	schedule.Post("/:id?", middleware.Protected(), scheduleHandler.Upsert)
}
