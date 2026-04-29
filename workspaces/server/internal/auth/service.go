package auth

import (
	"context"
	"fmt"

	"github.com/NishLy/go-fiber-boilerplate/config"
	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"github.com/NishLy/go-fiber-boilerplate/internal/token"
	"github.com/NishLy/go-fiber-boilerplate/internal/user"
	hash "github.com/NishLy/go-fiber-boilerplate/pkg/hash"
	jwt "github.com/NishLy/go-fiber-boilerplate/pkg/jwt"
)

type AuthServiceInterface interface {
	Login(ctx context.Context, email, password string) (string, string, error)
	Register(ctx context.Context, req RegisterRequest) (string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
}

type authService struct {
	userService  user.UserService
	tokenService token.TokenService
}

func NewAuthService(userService user.UserService, tokenService token.TokenService) *authService {
	return &authService{
		userService:  userService,
		tokenService: tokenService,
	}
}

func (j *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := j.userService.GetUserFromEmail(ctx, email)
	if err != nil {
		return "", "", apperror.UnauthorizedErr(err)
	}

	if !hash.CheckPasswordHash(password, user.Password) {
		return "", "", apperror.UnauthorizedErr(fmt.Errorf("invalid credentials"), "Invalid email or password")
	}

	tokenStr, err := j.tokenService.GenerateAccessToken(ctx, user.ID.String())
	if err != nil {
		return "", "", err
	}

	refreshTokenStr, err := j.tokenService.GenerateRefreshToken(ctx, user.ID.String())
	if err != nil {
		return "", "", err
	}

	return tokenStr, refreshTokenStr, nil
}

func (j *authService) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	user := domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	err := j.userService.RegisterUser(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (j *authService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	userId, err := jwt.VerifyToken(refreshToken, config.Get().JWT_SECRET, token.TokenTypeRefresh)

	if err != nil {
		return "", apperror.UnauthorizedErr(err, "Invalid refresh token")
	}

	// validate that the refresh token exists in the database
	storedToken, err :=
		j.tokenService.GetTokenFromUserID(ctx, userId, token.TokenTypeRefresh)

	if err != nil {
		return "", apperror.UnauthorizedErr(err, "Invalid refresh token")
	}

	if storedToken != refreshToken {
		return "", apperror.UnauthorizedErr(fmt.Errorf("refresh token mismatch"), "Invalid refresh token")
	}

	return j.tokenService.GenerateAccessToken(ctx, userId)
}
