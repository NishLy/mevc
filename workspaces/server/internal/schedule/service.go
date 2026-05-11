package schedule

import (
	"context"

	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
)

type ScheduleService interface {
	Upsert(ctx context.Context, id *uint64, schedule *UpsertScheduleRequest) (domain.Schedule, error)
	Delete(ctx context.Context, id string) error
}

type scheduleService struct {
	repo ScheduleRepository
}

func NewScheduleService(repo ScheduleRepository) ScheduleService {
	return &scheduleService{
		repo: repo,
	}
}

func (s *scheduleService) Upsert(ctx context.Context, id *uint64, schedule *UpsertScheduleRequest) (domain.Schedule, error) {
	if id != nil {
		schedule.ID = *id
	}
	return s.repo.Upsert(ctx, schedule)
}

func (s *scheduleService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
