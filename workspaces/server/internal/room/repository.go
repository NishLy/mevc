package room

import (
	"context"

	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/database"
	"go.uber.org/zap"
)

type RoomRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Room, error)
	FindByCode(ctx context.Context, code string) (*domain.Room, error)
	Upsert(ctx context.Context, room *domain.Room) (domain.Room, error)
	Delete(ctx context.Context, id string) error
}

type roomRepository struct {
	log *zap.SugaredLogger
}

func NewRoomRepository(log zap.SugaredLogger) RoomRepository {
	return &roomRepository{
		log: &log,
	}
}

func (r *roomRepository) FindByID(ctx context.Context, id string) (*domain.Room, error) {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return nil, database.Wrap(err)
	}

	var room domain.Room
	err = db.DB.Where("id = ?", id).Preload("Schedules").First(&room).Error
	if err != nil {
		return nil, database.Wrap(err)
	}

	return &room, nil
}
func (r *roomRepository) FindByCode(ctx context.Context, code string) (*domain.Room, error) {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return nil, database.Wrap(err)
	}

	var room domain.Room
	err = db.DB.Where("code = ?", code).Preload("Schedules").First(&room).Error
	if err != nil {
		return nil, database.Wrap(err)
	}

	return &room, nil
}

func (r *roomRepository) Upsert(ctx context.Context, room *domain.Room) (domain.Room, error) {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return domain.Room{}, database.Wrap(err)
	}

	err = db.DB.Save(room).Error

	if err != nil {
		return domain.Room{}, database.Wrap(err)
	}

	return *room, nil
}

func (r *roomRepository) Delete(ctx context.Context, id string) error {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return database.Wrap(err)
	}
	err = db.DB.Where("id = ?", id).Delete(&domain.Room{}).Error
	if err != nil {
		return database.Wrap(err)
	}

	return nil
}
