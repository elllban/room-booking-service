package repository

import (
	"database/sql"
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
)

type BookingRepository interface {
	Create(booking *domain.Booking) error
	GetByID(id uuid.UUID) (*domain.Booking, error)
	GetBySlotID(slotID uuid.UUID) (*domain.Booking, error)
	UpdateStatus(id uuid.UUID, status string) error
	ListAll(offset, limit int) ([]domain.Booking, int, error)
	ListByUserID(userID uuid.UUID, futureOnly bool) ([]domain.Booking, error)
	GetActiveBookingBySlotID(slotID uuid.UUID) (*domain.Booking, error)
}

type bookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(booking *domain.Booking) error {
	query := `
        INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	now := time.Now()
	_, err := r.db.Exec(query, booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink, now)
	if err != nil {
		return err
	}
	booking.CreatedAt = &now
	return nil
}

func (r *bookingRepository) GetByID(id uuid.UUID) (*domain.Booking, error) {
	query := `SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings WHERE id = $1`
	var booking domain.Booking
	err := r.db.QueryRow(query, id).Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) GetBySlotID(slotID uuid.UUID) (*domain.Booking, error) {
	query := `SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings WHERE slot_id = $1 AND status = 'active'`
	var booking domain.Booking
	err := r.db.QueryRow(query, slotID).Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) UpdateStatus(id uuid.UUID, status string) error {
	query := `UPDATE bookings SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *bookingRepository) ListAll(offset, limit int) ([]domain.Booking, int, error) {
	countQuery := `SELECT COUNT(*) FROM bookings`
	var total int
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
        SELECT id, slot_id, user_id, status, conference_link, created_at 
        FROM bookings 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var booking domain.Booking
		err = rows.Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, total, nil
}

func (r *bookingRepository) ListByUserID(userID uuid.UUID, futureOnly bool) ([]domain.Booking, error) {
	query := `
        SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at
        FROM bookings b
        JOIN slots s ON b.slot_id = s.id
        WHERE b.user_id = $1
    `
	if futureOnly {
		query += ` AND s.start_time >= NOW()`
	}
	query += ` ORDER BY s.start_time`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var booking domain.Booking
		err = rows.Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (r *bookingRepository) GetActiveBookingBySlotID(slotID uuid.UUID) (*domain.Booking, error) {
	query := `SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings WHERE slot_id = $1 AND status = 'active'`
	var booking domain.Booking
	err := r.db.QueryRow(query, slotID).Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &booking, nil
}
