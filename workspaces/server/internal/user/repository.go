package user

import (
	"context"

	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	fga "github.com/NishLy/go-fiber-boilerplate/internal/openfga"
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/database"
	"github.com/NishLy/go-fiber-boilerplate/internal/request"
	pkg "github.com/NishLy/go-fiber-boilerplate/pkg/context"
	"github.com/openfga/go-sdk/client"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
	GetUsers(ctx context.Context, pagination request.PaginationRequest) ([]domain.User, paginator.Cursor, error)
}

type userRepository struct {
	logger zap.SugaredLogger
}

func NewUserRepository(log zap.SugaredLogger) UserRepository {
	return &userRepository{logger: log}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return database.Wrap(err)
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		fgaClient, err := fga.GetFGAFromContext(ctx)
		if err != nil {
			r.logger.Errorf("Failed to get OpenFGA client from context: %v", err)
			return apperror.InternalErr(err)
		}

		err = tx.Create(user).Error

		if err != nil {
			r.logger.Errorf("Failed to create user: %v", err)
			return database.Wrap(err)
		}

		err = tx.Where("email = ?", user.Email).First(&user).Error
		if err != nil {
			r.logger.Errorf("Failed to get user by email: %v", err)
			return database.Wrap(err)
		}

		body := client.ClientWriteRequest{
			Writes: []client.ClientTupleKey{
				{
					User:     "user:" + user.ID.String(),
					Relation: "owner",
					Object:   "user:" + user.ID.String(),
				},
			},
		}

		tenantID := database.GetIndentifier(ctx)

		_, err = fgaClient.Write(context.Background()).
			Body(body).
			Options(*fga.GetFGAWriteOptions(fgaClient, tenantID)).
			Execute()

		if err != nil {
			r.logger.Errorf("Failed to write relationships to OpenFGA: %v", err)
			return apperror.InternalErr(err)
		}

		return nil
	})

	if err != nil {
		r.logger.Errorf("Failed to create user: %v", err)
		return database.Wrap(err)
	}

	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return nil, database.Wrap(err)
	}

	var user domain.User
	err = db.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		r.logger.Errorf("Failed to get user by email: %v", err)
		return nil, database.Wrap(err)
	}

	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return nil, database.Wrap(err)
	}

	var user domain.User
	err = db.DB.First(&user, id).Error
	if err != nil {
		r.logger.Errorf("Failed to get user by ID: %v", err)
		return nil, database.Wrap(err)
	}

	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return database.Wrap(err)
	}

	err = db.DB.Save(user).Error
	if err != nil {
		r.logger.Errorf("Failed to update user: %v", err)
		return database.Wrap(err)
	}
	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return database.Wrap(err)
	}
	err = db.DB.Delete(&domain.User{}, id).Error
	if err != nil {
		r.logger.Errorf("Failed to delete user: %v", err)
		return database.Wrap(err)
	}
	return nil
}

func (r *userRepository) GetUsers(ctx context.Context, pagination request.PaginationRequest) ([]domain.User, paginator.Cursor, error) {
	var users []domain.User

	db, err := database.GetDBFromContext(ctx)
	if err != nil {
		return nil, paginator.Cursor{}, database.Wrap(err)
	}

	p := paginator.New(&paginator.Config{
		// clean up the sort_by input to prevent SQL injection
		Keys:  []string{domain.GetSortColumn(pagination.SortBy)},
		Limit: pagination.Limit,
		Order: paginator.Order(pagination.Sort),
	})

	if pagination.AfterCursor != "" {
		p.SetAfterCursor(pagination.AfterCursor)
	}

	fgaClient, err := fga.GetFGAFromContext(ctx)

	if err != nil {
		r.logger.Errorf("Failed to get OpenFGA client from context: %v", err)
		return nil, paginator.Cursor{}, apperror.InternalErr(err)
	}

	userID, err := pkg.GetSubFromContext(ctx)
	if err != nil {
		r.logger.Errorf("Failed to get user ID from context: %v", err)
		return nil, paginator.Cursor{}, apperror.InternalErr(err)
	}

	body := client.ClientCheckRequest{
		User:     "user:" + userID,
		Relation: "admin",
		Object:   "system:main",
	}

	tenantID := database.GetIndentifier(ctx)

	data, err := fgaClient.Check(context.Background()).
		Body(body).
		Options(*fga.GetFGAClientCheckOptions(fgaClient, tenantID)).
		Execute()

	if err != nil {
		r.logger.Errorf("Failed to check relationship in OpenFGA: %v", err)
		return nil, paginator.Cursor{}, apperror.InternalErr(err)
	}

	query := db.Model(&domain.User{})

	if pagination.Search != "" {
		searchTerm := "%" + pagination.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
	}

	if data.Allowed != nil && !*data.Allowed {
		query = query.Where("id = ?", userID)
	}

	result, cursor, err := p.Paginate(query, &users)
	if err != nil {
		return nil, paginator.Cursor{}, database.Wrap(err)
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, database.Wrap(result.Error)
	}

	return users, cursor, nil
}
