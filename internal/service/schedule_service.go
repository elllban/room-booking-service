package service

import (
	"errors"
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/elllban/test-backend-elllban/internal/repository"
	"github.com/google/uuid"
)

type ScheduleService interface {
	CreateSchedule(roomID uuid.UUID, req *domain.CreateScheduleRequest) (*domain.Schedule, error)
	GenerateSlotsForDate(roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
}

type scheduleService struct {
	scheduleRepo repository.ScheduleRepository
	roomRepo     repository.RoomRepository
	slotRepo     repository.SlotRepository
}

func NewScheduleService(scheduleRepo repository.ScheduleRepository, roomRepo repository.RoomRepository, slotRepo repository.SlotRepository) ScheduleService {
	return &scheduleService{
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
		slotRepo:     slotRepo,
	}
}

func (s *scheduleService) CreateSchedule(roomID uuid.UUID, req *domain.CreateScheduleRequest) (*domain.Schedule, error) {
	exists, err := s.roomRepo.Exists(roomID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("room not found")
	}

	exists, err = s.scheduleRepo.ExistsForRoom(roomID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("schedule already exists for this room")
	}

	for _, day := range req.DaysOfWeek {
		if day < 1 || day > 7 {
			return nil, errors.New("days of week must be between 1 and 7")
		}
	}

	if _, err = time.Parse("15:04", req.StartTime); err != nil {
		return nil, errors.New("invalid start time format, expected HH:MM")
	}
	if _, err = time.Parse("15:04", req.EndTime); err != nil {
		return nil, errors.New("invalid end time format, expected HH:MM")
	}

	schedule := &domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: req.DaysOfWeek,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	}

	if err = s.scheduleRepo.Create(schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

func (s *scheduleService) GenerateSlotsForDate(roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	schedule, err := s.scheduleRepo.GetByRoomID(roomID)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return []domain.Slot{}, nil
	}

	weekday := int(date.Weekday())
	var ourWeekday int
	if weekday == 0 {
		ourWeekday = 7
	} else {
		ourWeekday = weekday
	}

	isAvailable := false
	for _, day := range schedule.DaysOfWeek {
		if day == ourWeekday {
			isAvailable = true
			break
		}
	}
	if !isAvailable {
		return []domain.Slot{}, nil
	}

	startTime, _ := time.Parse("15:04", schedule.StartTime)
	endTime, _ := time.Parse("15:04", schedule.EndTime)

	baseDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	startDateTime := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	endDateTime := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	var slots []domain.Slot
	current := startDateTime
	for current.Before(endDateTime) {
		slotEnd := current.Add(30 * time.Minute)
		if slotEnd.After(endDateTime) {
			break
		}

		slot := &domain.Slot{
			ID:     uuid.New(),
			RoomID: roomID,
			Start:  current,
			End:    slotEnd,
		}

		if err = s.slotRepo.Create(slot); err == nil {
			slots = append(slots, *slot)
		} else {
			existingSlots, _ := s.slotRepo.GetSlotsForDate(roomID, date)
			for _, existing := range existingSlots {
				if existing.Start.Equal(current) {
					slots = append(slots, existing)
					break
				}
			}
		}

		current = slotEnd
	}

	return slots, nil
}
