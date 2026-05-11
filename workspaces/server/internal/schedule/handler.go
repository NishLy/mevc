package schedule

import (
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"github.com/NishLy/go-fiber-boilerplate/internal/response"
	pkg "github.com/NishLy/go-fiber-boilerplate/pkg/fiber"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/NishLy/go-fiber-boilerplate/pkg/validator"
	"github.com/gofiber/fiber/v3"
)

type ScheduleHandler interface {
	Upsert(c fiber.Ctx) error
}

type scheduleHandler struct {
	scheduleService ScheduleService
}

func NewScheduleHandler(scheduleService ScheduleService) ScheduleHandler {
	return &scheduleHandler{scheduleService: scheduleService}
}

// Upsert godoc
// @Summary Upsert schedule
// @Description Create or update a schedule
// @Tags schedule
// @Accept json
// @Produce json
// @Param request body UpsertScheduleRequest true "Create or update schedule request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /schedule/upsert [post]
func (r *scheduleHandler) Upsert(c fiber.Ctx) error {
	id, err := pkg.GetUint64PtrFromParams(c, "id")

	var req UpsertScheduleRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Sugar.Errorf("Failed to bind request body: %v", err)
		return apperror.BadRequestErr(err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := apperror.ParseValidationErrors(err)
		return apperror.ValidationErr(validationErrors)
	}

	_, err = r.scheduleService.Upsert(c.Context(), id, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).
		JSON(response.GenericResponse{
			Code:    fiber.StatusCreated,
			Message: "Schedule upserted successfully",
		})
}
