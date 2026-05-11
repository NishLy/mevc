package room

import (
	"context"

	"github.com/NishLy/go-fiber-boilerplate/internal/domain"
	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"github.com/NishLy/go-fiber-boilerplate/pkg/validator"
	"github.com/google/uuid"
)

type RoomService interface {
	GetRoomByID(ctx context.Context, id string) (*domain.Room, error)
	GetRoomByCode(ctx context.Context, code string) (*domain.Room, error)
	UpsertRoom(ctx context.Context, id *uint64, room *CreateRoomRequest) (domain.Room, error)
	DeleteRoom(ctx context.Context, id string) error
}

type roomService struct {
	repo RoomRepository
}

func NewRoomService(repo RoomRepository) RoomService {
	return &roomService{
		repo: repo,
	}
}

func (s *roomService) GetRoomByCode(ctx context.Context, code string) (*domain.Room, error) {
	room, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *roomService) GetRoomByID(ctx context.Context, id string) (*domain.Room, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *roomService) UpsertRoom(ctx context.Context, id *uint64, room *CreateRoomRequest) (domain.Room, error) {

	userID := ctx.Value("user_id").(string)
	uuid, err := uuid.Parse(userID)
	if userID == "" || err != nil {
		return domain.Room{}, apperror.UnauthorizedErr(nil, "User ID not found in context")
	}

	roomDomain := &domain.Room{
		Name:             room.Name,
		Description:      room.Description,
		HostID:           uuid,
		AutoJoin:         *validator.AssignOrDefault(room.Settings.AutoJoin, true),
		AllowGuests:      *validator.AssignOrDefault(room.Settings.AllowGuests, true),
		AllowRecording:   *validator.AssignOrDefault(room.Settings.AllowRecording, true),
		AllowChat:        *validator.AssignOrDefault(room.Settings.AllowChat, true),
		AllowScreenShare: *validator.AssignOrDefault(room.Settings.AllowScreenShare, true),
		Capacity:         *validator.AssignOrDefault(room.Settings.Capacity, uint(10)),
		Location:         *validator.AssignOrDefault(room.Settings.Location, "remote"),
	}

	if id != nil {
		roomDomain.ID = *id
	}

	return s.repo.Upsert(ctx, roomDomain)
}

func (s *roomService) DeleteRoom(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
