package service

import (
	"errors"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/elllban/test-backend-elllban/internal/repository"
	"github.com/google/uuid"
)

type RoomService interface {
	CreateRoom(req *domain.CreateRoomRequest) (*domain.Room, error)
	ListRooms() ([]domain.Room, error)
}

type roomService struct {
	roomRepo repository.RoomRepository
}

func NewRoomService(roomRepo repository.RoomRepository) RoomService {
	return &roomService{
		roomRepo: roomRepo,
	}
}

func (s *roomService) CreateRoom(req *domain.CreateRoomRequest) (*domain.Room, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	room := &domain.Room{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *roomService) ListRooms() ([]domain.Room, error) {
	return s.roomRepo.List()
}
