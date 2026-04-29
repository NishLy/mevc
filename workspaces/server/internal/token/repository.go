package token

import (
	"context"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/database"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TokenRepository interface {
	SaveToken(ctx context.Context, userID string, token string, tokenType string, expires time.Duration) error
	DeleteTokenByUserID(ctx context.Context, userID string, tokenType string) error
	DeleteAllTokensByUserID(ctx context.Context, userID string) error
	GetTokenByUserIDAndType(ctx context.Context, userID string, tokenType string) (*domain.Token, error)
	GetTokenFromUserID(ctx context.Context, userID string, tokenType string) (string, error)
	CleanupExpiredTokens(ctx context.Context) error
}

type tokenRepository struct {
	logger zap.SugaredLogger
}

func NewTokenRepository(logger zap.SugaredLogger) TokenRepository {
	return &tokenRepository{
		logger: logger,
	}
}

func (r *tokenRepository) SaveToken(ctx context.Context, userID string, token string, tokenType string, expires time.Duration) error {
	db, err := database.GetDBFromContext(ctx)
	if err != nil {
		return database.Wrap(err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		tokenRecord := &domain.Token{
			ID:      uuid.New(),
			Token:   token,
			UserID:  uuid.MustParse(userID),
			Type:    tokenType,
			Expires: time.Now().Add(expires),
		}

		contextWithTx := database.SetContextWithTx(ctx, tx)

		DeleteTokenByUserIDErr := r.DeleteTokenByUserID(contextWithTx, userID, tokenType)
		if DeleteTokenByUserIDErr != nil {
			r.logger.Errorf("Failed to delete token by user ID: %v", DeleteTokenByUserIDErr)
			return database.Wrap(DeleteTokenByUserIDErr)
		}

		err = tx.Create(tokenRecord).Error
		if err != nil {
			r.logger.Errorf("Failed to save token: %v", err)
			return database.Wrap(err)
		}

		return nil
	})

	if err != nil {
		return database.Wrap(err)
	}

	return nil
}

func (r *tokenRepository) DeleteTokenByUserID(ctx context.Context, userID string, tokenType string) error {
	db, err := database.GetDBFromContext(ctx)
	if err != nil {
		return database.Wrap(err)
	}

	err = db.Where("user_id = ? AND type = ?", userID, tokenType).Delete(&domain.Token{}).Error
	if err != nil {
		r.logger.Errorf("Failed to delete token: %v", err)
		return database.Wrap(err)
	}

	return nil
}

func (r *tokenRepository) DeleteAllTokensByUserID(ctx context.Context, userID string) error {
	db, err := database.GetDBFromContext(ctx)
	if err != nil {
		return database.Wrap(err)
	}
	err = db.Where("user_id = ?", userID).Delete(&domain.Token{}).Error
	if err != nil {
		r.logger.Errorf("Failed to delete all tokens: %v", err)
		return database.Wrap(err)
	}

	return nil
}

func (r *tokenRepository) GetTokenByUserIDAndType(ctx context.Context, userID string, tokenType string) (*domain.Token, error) {
	db, err := database.GetDBFromContext(ctx)
	if err != nil {
		return nil, database.Wrap(err)
	}
	var token domain.Token
	err = db.Where("user_id = ? AND type = ?", userID, tokenType).First(&token).Error
	if err != nil {
		r.logger.Errorf("Failed to get token: %v", err)
		return nil, database.Wrap(err)
	}

	return &token, nil
}

func (r *tokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	db, err := database.GetDBFromContext(ctx)
	if err != nil {
		return database.Wrap(err)
	}

	err = db.Where("expires < ?", time.Now()).Delete(&domain.Token{}).Error
	if err != nil {
		r.logger.Errorf("Failed to cleanup expired tokens: %v", err)
		return database.Wrap(err)
	}

	return nil
}

func (r *tokenRepository) GetTokenFromUserID(ctx context.Context, userID string, tokenType string) (string, error) {
	tokenRecord, err := r.GetTokenByUserIDAndType(ctx, userID, tokenType)
	if err != nil {
		return "", err
	}

	return tokenRecord.Token, nil
}
