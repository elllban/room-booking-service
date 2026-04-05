package service

import (
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/elllban/test-backend-elllban/internal/repository"
	"github.com/google/uuid"
)

type SlotService interface {
	GetAvailableSlots(roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
	GetSlotByID(id uuid.UUID) (*domain.Slot, error)
}

type slotService struct {
	slotRepo     repository.SlotRepository
	scheduleRepo repository.ScheduleRepository
}

func NewSlotService(slotRepo repository.SlotRepository, scheduleRepo repository.ScheduleRepository) SlotService {
	return &slotService{
		slotRepo:     slotRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (s *slotService) GetAvailableSlots(roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	schedule, err := s.scheduleRepo.GetByRoomID(roomID)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return []domain.Slot{}, nil
	}

	slots, err := s.slotRepo.GetSlotsForDate(roomID, date)
	if err != nil {
		return nil, err
	}

	if len(slots) == 0 {
		return []domain.Slot{}, nil
	}

	var slotIDs []uuid.UUID
	for _, slot := range slots {
		slotIDs = append(slotIDs, slot.ID)
	}

	bookedSlots, err := s.slotRepo.GetBookedSlotIDs(slotIDs)
	if err != nil {
		return nil, err
	}

	var availableSlots []domain.Slot
	for _, slot := range slots {
		if !bookedSlots[slot.ID] {
			availableSlots = append(availableSlots, slot)
		}
	}

	return availableSlots, nil
}

func (s *slotService) GetSlotByID(id uuid.UUID) (*domain.Slot, error) {
	return s.slotRepo.GetByID(id)
}
