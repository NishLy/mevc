package schedule

import (
	"context"
	"encoding/json"

	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/database"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type ScheduleRepository interface {
	Upsert(ctx context.Context, schedule *UpsertScheduleRequest) (domain.Schedule, error)
	Delete(ctx context.Context, id string) error
}

type scheduleRepository struct {
	log *zap.SugaredLogger
}

func NewScheduleRepository(log zap.SugaredLogger) ScheduleRepository {
	return &scheduleRepository{
		log: &log,
	}
}

func (r *scheduleRepository) Upsert(ctx context.Context, schedule *UpsertScheduleRequest) (domain.Schedule, error) {
	parsedJson, err := json.Marshal(schedule.Pattern)
	if err != nil {
		return domain.Schedule{}, err
	}

	secheduleDomain := domain.Schedule{
		ID:      schedule.ID,
		RoomID:  schedule.RoomID,
		Start:   schedule.Start,
		End:     schedule.End,
		Pattern: datatypes.JSON(parsedJson),
	}

	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return domain.Schedule{}, database.Wrap(err)
	}

	err = db.DB.Save(&secheduleDomain).Error

	if err != nil {
		return domain.Schedule{}, database.Wrap(err)
	}

	return secheduleDomain, nil
}

func (r *scheduleRepository) Delete(ctx context.Context, id string) error {
	db, err := database.GetDB(database.GetIndentifier(ctx), false)
	if err != nil {
		return database.Wrap(err)
	}
	err = db.DB.Where("id = ?", id).Delete(&domain.Schedule{}).Error
	if err != nil {
		return database.Wrap(err)
	}

	return nil
}
