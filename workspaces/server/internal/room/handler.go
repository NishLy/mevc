package room

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"github.com/NishLy/go-fiber-boilerplate/internal/response"
	pkg "github.com/NishLy/go-fiber-boilerplate/pkg/fiber"
	"github.com/NishLy/go-fiber-boilerplate/pkg/validator"
	"github.com/gofiber/fiber/v3"
)

type RoomHandler interface {
	GetRoomByID(c fiber.Ctx) error
	GetRoomByCode(c fiber.Ctx) error
	Upsert(c fiber.Ctx) error
	Delete(c fiber.Ctx) error
}

type roomHandler struct {
	roomService RoomService
}

func NewRoomHandler(roomService RoomService) RoomHandler {
	return &roomHandler{roomService: roomService}
}

// FindByID godoc
// @Summary Get room by ID
// @Description Retrieve a room by its ID
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/login [post]
func (r *roomHandler) GetRoomByID(c fiber.Ctx) error {
	id := c.Params("id")

	room, err := r.roomService.GetRoomByID(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse[*domain.Room]{
		GenericResponse: response.GenericResponse{
			Code:    fiber.StatusOK,
			Message: "Room retrieved successfully",
		},
		Data: room,
	})
}

// FindByCode godoc
// @Summary Get room by code
// @Description Retrieve a room by its code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/login [post]
func (r *roomHandler) GetRoomByCode(c fiber.Ctx) error {
	code := c.Params("code")

	room, err := r.roomService.GetRoomByCode(c.Context(), code)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.GenericSuccessResponse[*domain.Room]{
		GenericResponse: response.GenericResponse{
			Code:    fiber.StatusOK,
			Message: "Room retrieved successfully",
		},
		Data: room,
	})
}

// Upsert godoc
// @Summary Upsert room
// @Description Create or update a room
// @Tags room
// @Accept json
// @Produce json
// @Param request body CreateRoomRequest true "Create or update room request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /room/upsert [post]
func (r *roomHandler) Upsert(c fiber.Ctx) error {
	id, err := pkg.GetUint64PtrFromParams(c, "id")

	var req CreateRoomRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperror.BadRequestErr(err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := apperror.ParseValidationErrors(err)
		return apperror.ValidationErr(validationErrors)
	}

	_, err = r.roomService.UpsertRoom(c.Context(), id, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).
		JSON(response.GenericResponse{
			Code:    fiber.StatusCreated,
			Message: "Room upserted successfully",
		})
}

// Delete godoc
// @Summary Delete room
// @Description Delete a room by its ID
// @Tags room
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /room/{id} [delete]
func (r *roomHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")

	err := r.roomService.DeleteRoom(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(response.GenericResponse{
			Code:    fiber.StatusOK,
			Message: "Room deleted successfully",
		})
}
