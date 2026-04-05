package repository

import (
	"database/sql"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ScheduleRepository interface {
	Create(schedule *domain.Schedule) error
	GetByRoomID(roomID uuid.UUID) (*domain.Schedule, error)
	ExistsForRoom(roomID uuid.UUID) (bool, error)
}

type scheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) Create(schedule *domain.Schedule) error {
	query := `
        INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.Exec(query, schedule.ID, schedule.RoomID, pq.Array(schedule.DaysOfWeek), schedule.StartTime, schedule.EndTime)
	return err
}

func (r *scheduleRepository) GetByRoomID(roomID uuid.UUID) (*domain.Schedule, error) {
	query := `SELECT id, room_id, days_of_week, start_time, end_time FROM schedules WHERE room_id = $1`
	var schedule domain.Schedule
	var daysOfWeek []int
	err := r.db.QueryRow(query, roomID).Scan(&schedule.ID, &schedule.RoomID, pq.Array(&daysOfWeek), &schedule.StartTime, &schedule.EndTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	schedule.DaysOfWeek = daysOfWeek
	return &schedule, nil
}

func (r *scheduleRepository) ExistsForRoom(roomID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM schedules WHERE room_id = $1)`
	var exists bool
	err := r.db.QueryRow(query, roomID).Scan(&exists)
	return exists, err
}
