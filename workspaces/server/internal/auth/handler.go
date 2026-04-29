package auth

import (
	"github.com/NishLy/go-fiber-boilerplate/config"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"github.com/NishLy/go-fiber-boilerplate/internal/response"
	"github.com/NishLy/go-fiber-boilerplate/pkg/validator"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler interface {
	Login(c fiber.Ctx) error
	Register(c fiber.Ctx) error
	RefreshToken(c fiber.Ctx) error
}

type authHandler struct {
	authService *authService
}

func NewAuthHandler(authService *authService) AuthHandler {
	return &authHandler{authService: authService}
}

// Login godoc
// @Summary User login
// @Description Authenticate user with credentials
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/login [post]
func (a *authHandler) Login(c fiber.Ctx) error {

	var req LoginRequest

	if err := c.Bind().Body(&req); err != nil {
		return apperror.BadRequestErr(err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := apperror.ParseValidationErrors(err)
		return apperror.ValidationErr(validationErrors)
	}

	token, refreshToken, err := a.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	cfg := config.Get()

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   cfg.ENV == "production",
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).
		JSON(response.GenericSuccessResponse[fiber.Map]{
			GenericResponse: response.GenericResponse{
				Code:    fiber.StatusOK,
				Message: "Login successful",
			},
			Data: fiber.Map{
				"token":        token,
				"refreshToken": refreshToken,
			},
		})
}

// Register godoc
// @Summary User registration
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (a *authHandler) Register(c fiber.Ctx) error {

	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperror.BadRequestErr(err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := apperror.ParseValidationErrors(err)
		return apperror.ValidationErr(validationErrors)
	}

	_, err := a.authService.Register(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).
		JSON(response.GenericResponse{
			Code:    fiber.StatusCreated,
			Message: "Registration successful",
		})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh the access token using the refresh token stored in the cookie
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (a *authHandler) RefreshToken(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken == "" {
		return apperror.UnauthorizedErr(nil, "No refresh token provided")
	}

	newToken, err := a.authService.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   config.Get().ENV == "production",
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).
		JSON(response.GenericSuccessResponse[fiber.Map]{
			GenericResponse: response.GenericResponse{
				Code:    fiber.StatusOK,
				Message: "Token refreshed successfully",
			},
			Data: fiber.Map{
				"token": newToken,
			},
		})
}
