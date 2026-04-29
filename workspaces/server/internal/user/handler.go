package user

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"github.com/NishLy/go-fiber-boilerplate/internal/request"
	"github.com/NishLy/go-fiber-boilerplate/internal/response"
	"github.com/NishLy/go-fiber-boilerplate/pkg/validator"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type UserHandler interface {
	GetUsers(c fiber.Ctx) error
}

type userHandler struct {
	logger      *zap.SugaredLogger
	userService UserService
}

func NewUserHandler(logger *zap.SugaredLogger, userService UserService) UserHandler {
	return &userHandler{
		logger:      logger,
		userService: userService,
	}
}

// GetUsers retrieves a paginated list of users.
//
//		@Summary      Get users
//		@Description  Retrieve a paginated list of all users
//		@Tags         users
//		@Produce      json
//		@Param        before  query     string  false  "Cursor before"
//		@Param        after   query     string  false  "Cursor after"
//		@Param        limit   query     int     false  "Limit"
//	 @Param		  search  query     string  false  "Search term"
//		@Success      200     {object}  response.PagedDataResponse[domain.User]
//		@Failure      400     {object}  apperror.Error
//		@Router       /users [get]
func (h *userHandler) GetUsers(c fiber.Ctx) error {
	var req request.PaginationRequest

	if err := c.Bind().Query(&req); err != nil {
		return apperror.BadRequestErr(err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		return apperror.BadRequestErr(err)
	}

	users, cursor, err := h.userService.GetUsers(c.Context(), req)

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.PagedDataResponse[domain.User]{
		GenericResponse: response.GenericResponse{
			Code:    fiber.StatusOK,
			Message: "Users retrieved successfully",
		},
		Data: users,
		Meta: response.PaginationMeta{
			Before:  cursor.Before,
			After:   cursor.After,
			HasNext: cursor.After != nil,
			HasPrev: cursor.Before != nil,
		},
	})
}

// fiber:context-methods migrated
