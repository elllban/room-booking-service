package service

import (
	"testing"
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
)

type mockScheduleRepoForSlot struct {
	schedule *domain.Schedule
	exists   bool
}

func (m *mockScheduleRepoForSlot) Create(schedule *domain.Schedule) error {
	return nil
}

func (m *mockScheduleRepoForSlot) GetByRoomID(roomID uuid.UUID) (*domain.Schedule, error) {
	if m.exists {
		return m.schedule, nil
	}
	return nil, nil
}

func (m *mockScheduleRepoForSlot) ExistsForRoom(roomID uuid.UUID) (bool, error) {
	return m.exists, nil
}

type mockSlotRepoForSlot struct {
	slots []domain.Slot
}

func (m *mockSlotRepoForSlot) Create(slot *domain.Slot) error {
	return nil
}

func (m *mockSlotRepoForSlot) GetByID(id uuid.UUID) (*domain.Slot, error) {
	return nil, nil
}

func (m *mockSlotRepoForSlot) GetAvailableSlots(roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	return nil, nil
}

func (m *mockSlotRepoForSlot) GetBookedSlotIDs(slotIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return make(map[uuid.UUID]bool), nil
}

func (m *mockSlotRepoForSlot) GetSlotsForDate(roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	return m.slots, nil
}

func (m *mockSlotRepoForSlot) DeleteSlotsForRoom(roomID uuid.UUID) error {
	return nil
}

func TestGetAvailableSlots_NoSchedule(t *testing.T) {
	slotRepo := &mockSlotRepoForSlot{}
	scheduleRepo := &mockScheduleRepoForSlot{exists: false} // Нет расписания
	service := NewSlotService(slotRepo, scheduleRepo)

	slots, err := service.GetAvailableSlots(uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(slots) != 0 {
		t.Errorf("Expected empty slots, got %d", len(slots))
	}
}

func TestGetAvailableSlots_WithSchedule(t *testing.T) {
	slotRepo := &mockSlotRepoForSlot{
		slots: []domain.Slot{
			{ID: uuid.New(), Start: time.Now(), End: time.Now().Add(30 * time.Minute)},
		},
	}
	scheduleRepo := &mockScheduleRepoForSlot{
		exists: true,
		schedule: &domain.Schedule{
			DaysOfWeek: []int{1, 2, 3, 4, 5},
			StartTime:  "09:00",
			EndTime:    "17:00",
		},
	}
	service := NewSlotService(slotRepo, scheduleRepo)

	slots, err := service.GetAvailableSlots(uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(slots) == 0 {
		t.Error("Expected non-empty slots, got empty")
	}
}
