package repository

import (
	"database/sql"
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type SlotRepository interface {
	Create(slot *domain.Slot) error
	GetByID(id uuid.UUID) (*domain.Slot, error)
	GetAvailableSlots(roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
	GetBookedSlotIDs(slotIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	GetSlotsForDate(roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
	DeleteSlotsForRoom(roomID uuid.UUID) error
}

type slotRepository struct {
	db *sql.DB
}

func NewSlotRepository(db *sql.DB) SlotRepository {
	return &slotRepository{db: db}
}

func (r *slotRepository) Create(slot *domain.Slot) error {
	query := `
        INSERT INTO slots (id, room_id, start_time, end_time)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (room_id, start_time) DO NOTHING
    `
	_, err := r.db.Exec(query, slot.ID, slot.RoomID, slot.Start, slot.End)
	return err
}

func (r *slotRepository) GetByID(id uuid.UUID) (*domain.Slot, error) {
	query := `SELECT id, room_id, start_time, end_time FROM slots WHERE id = $1`
	var slot domain.Slot
	err := r.db.QueryRow(query, id).Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *slotRepository) GetAvailableSlots(roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
        SELECT s.id, s.room_id, s.start_time, s.end_time
        FROM slots s
        LEFT JOIN bookings b ON s.id = b.slot_id AND b.status = 'active'
        WHERE s.room_id = $1 
          AND s.start_time >= $2 
          AND s.start_time < $3
          AND b.id IS NULL
        ORDER BY s.start_time
    `
	rows, err := r.db.Query(query, roomID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []domain.Slot
	for rows.Next() {
		var slot domain.Slot
		err = rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	return slots, nil
}

func (r *slotRepository) GetBookedSlotIDs(slotIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(slotIDs) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	query := `SELECT DISTINCT slot_id FROM bookings WHERE slot_id = ANY($1) AND status = 'active'`
	rows, err := r.db.Query(query, pq.Array(slotIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	booked := make(map[uuid.UUID]bool)
	for rows.Next() {
		var slotID uuid.UUID
		if err = rows.Scan(&slotID); err != nil {
			return nil, err
		}
		booked[slotID] = true
	}
	return booked, nil
}

func (r *slotRepository) GetSlotsForDate(roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
        SELECT id, room_id, start_time, end_time
        FROM slots
        WHERE room_id = $1 AND start_time >= $2 AND start_time < $3
        ORDER BY start_time
    `
	rows, err := r.db.Query(query, roomID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []domain.Slot
	for rows.Next() {
		var slot domain.Slot
		err = rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	return slots, nil
}

func (r *slotRepository) DeleteSlotsForRoom(roomID uuid.UUID) error {
	query := `DELETE FROM slots WHERE room_id = $1`
	_, err := r.db.Exec(query, roomID)
	return err
}
