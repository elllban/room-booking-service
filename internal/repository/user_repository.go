package repository

import (
	"database/sql"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByID(id uuid.UUID) (*domain.User, error)
	CreateOrGetByID(id uuid.UUID, role string) (*domain.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, email, role, created_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateOrGetByID(id uuid.UUID, role string) (*domain.User, error) {
	user, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	query := `INSERT INTO users (id, email, role) VALUES ($1, $2, $3)`
	email := id.String() + "@example.com" // Dummy email for test users
	_, err = r.db.Exec(query, id, email, role)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}
