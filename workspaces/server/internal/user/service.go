package user

import (
	"context"

	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	"github.com/NishLy/go-fiber-boilerplate/internal/request"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"go.uber.org/zap"
)

type UserService interface {
	GetUserFromEmail(ctx context.Context, email string) (*domain.User, error)
	RegisterUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error
	GetUsers(ctx context.Context, pagination request.PaginationRequest) ([]domain.User, paginator.Cursor, error)
}

type userService struct {
	repo   UserRepository
	logger zap.SugaredLogger
}

func NewUserService(repo UserRepository, logger zap.SugaredLogger) UserService {
	return &userService{repo: repo, logger: logger}
}

func (s *userService) GetUserFromEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *userService) RegisterUser(ctx context.Context, user *domain.User) error {
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.repo.UpdateUser(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.DeleteUser(ctx, userID)
}

func (s *userService) GetUsers(ctx context.Context, pagination request.PaginationRequest) ([]domain.User, paginator.Cursor, error) {
	return s.repo.GetUsers(ctx, pagination)
}
