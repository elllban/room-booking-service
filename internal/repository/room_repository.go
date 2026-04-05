package repository

import (
	"database/sql"
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(room *domain.Room) error
	List() ([]domain.Room, error)
	GetByID(id uuid.UUID) (*domain.Room, error)
	Exists(id uuid.UUID) (bool, error)
}

type roomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(room *domain.Room) error {
	query := `
        INSERT INTO rooms (id, name, description, capacity, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	now := time.Now()
	_, err := r.db.Exec(query, room.ID, room.Name, room.Description, room.Capacity, now)
	if err != nil {
		return err
	}
	room.CreatedAt = &now
	return nil
}

func (r *roomRepository) List() ([]domain.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		err = rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *roomRepository) GetByID(id uuid.UUID) (*domain.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`
	var room domain.Room
	err := r.db.QueryRow(query, id).Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) Exists(id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(query, id).Scan(&exists)
	return exists, err
}
