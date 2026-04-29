package token

import (
	"context"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/config"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	pkg "github.com/NishLy/go-fiber-boilerplate/pkg/jwt"
	"go.uber.org/zap"
)

type TokenService interface {
	GenerateToken(ctx context.Context, userID string, expires time.Duration, tokenType string) (string, error)
	GenerateRefreshToken(ctx context.Context, userID string) (string, error)
	GenerateAccessToken(ctx context.Context, userID string) (string, error)
	DeleteTokenByUserID(ctx context.Context, userID string, tokenType string) error
	GenerateForgotPasswordToken(ctx context.Context, userID string) (string, error)
	GetTokenFromUserID(ctx context.Context, userID string, tokenType string) (string, error)
}

type tokenService struct {
	logger zap.SugaredLogger
	r      TokenRepository
}

func NewTokenService(logger zap.SugaredLogger, r TokenRepository) TokenService {
	return &tokenService{
		logger: logger,
		r:      r,
	}
}

func (s *tokenService) GenerateToken(ctx context.Context, userID string, expires time.Duration, tokenType string) (string, error) {
	cfg := config.Get()

	tokenStr, err := pkg.GenerateToken(userID, cfg.JWT_SECRET, tokenType, int64(expires.Seconds()))

	if err != nil {
		s.logger.Errorf("Failed to generate token: %v", err)
		return "", apperror.InternalErr(err)
	}

	err = s.r.SaveToken(ctx, userID, tokenStr, tokenType, expires)
	if err != nil {
		s.logger.Errorf("Failed to save token: %v", err)
		return "", err
	}

	return tokenStr, nil
}

func (s *tokenService) GenerateRefreshToken(ctx context.Context, userID string) (string, error) {
	return s.GenerateToken(ctx, userID, time.Duration(config.Get().REFRESH_TOKEN_EXPIRATION)*time.Second, TokenTypeRefresh)
}

func (s *tokenService) GenerateAccessToken(ctx context.Context, userID string) (string, error) {
	return s.GenerateToken(ctx, userID, time.Duration(config.Get().ACCESS_TOKEN_EXPIRATION)*time.Second, TokenTypeAccess)
}

func (s *tokenService) DeleteTokenByUserID(ctx context.Context, userID string, tokenType string) error {
	return s.r.DeleteTokenByUserID(ctx, userID, tokenType)
}

func (s *tokenService) GenerateForgotPasswordToken(ctx context.Context, userID string) (string, error) {
	return s.GenerateToken(ctx, userID, time.Duration(config.Get().FORGOT_PASSWORD_TOKEN_EXPIRATION)*time.Second, TokenTypeResetPassword)
}

func (s *tokenService) GetTokenFromUserID(ctx context.Context, userID string, tokenType string) (string, error) {
	return s.r.GetTokenFromUserID(ctx, userID, tokenType)
}
