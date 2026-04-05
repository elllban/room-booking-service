package service

import (
	"testing"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
)

type mockScheduleRepo struct {
	schedules map[uuid.UUID]*domain.Schedule
}

func (m *mockScheduleRepo) Create(schedule *domain.Schedule) error {
	if m.schedules == nil {
		m.schedules = make(map[uuid.UUID]*domain.Schedule)
	}
	m.schedules[schedule.RoomID] = schedule
	return nil
}

func (m *mockScheduleRepo) GetByRoomID(roomID uuid.UUID) (*domain.Schedule, error) {
	if m.schedules == nil {
		return nil, nil
	}
	return m.schedules[roomID], nil
}

func (m *mockScheduleRepo) ExistsForRoom(roomID uuid.UUID) (bool, error) {
	if m.schedules == nil {
		return false, nil
	}
	_, exists := m.schedules[roomID]
	return exists, nil
}

func TestCreateSchedule_InvalidDays(t *testing.T) {
	scheduleRepo := &mockScheduleRepo{}
	roomRepo := &mockRoomRepo{}
	slotRepo := &mockSlotRepo{}
	service := NewScheduleService(scheduleRepo, roomRepo, slotRepo)

	roomID := uuid.New()
	roomRepo.Create(&domain.Room{ID: roomID, Name: "Test"})

	req := &domain.CreateScheduleRequest{
		DaysOfWeek: []int{8},
		StartTime:  "09:00",
		EndTime:    "17:00",
	}

	_, err := service.CreateSchedule(roomID, req)
	if err == nil {
		t.Error("Expected error for invalid day of week")
	}
}

func TestCreateSchedule_InvalidTime(t *testing.T) {
	scheduleRepo := &mockScheduleRepo{}
	roomRepo := &mockRoomRepo{}
	slotRepo := &mockSlotRepo{}
	service := NewScheduleService(scheduleRepo, roomRepo, slotRepo)

	roomID := uuid.New()
	roomRepo.Create(&domain.Room{ID: roomID, Name: "Test"})

	req := &domain.CreateScheduleRequest{
		DaysOfWeek: []int{1},
		StartTime:  "25:00",
		EndTime:    "17:00",
	}

	_, err := service.CreateSchedule(roomID, req)
	if err == nil {
		t.Error("Expected error for invalid time format")
	}
}
