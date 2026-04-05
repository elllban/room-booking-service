package service

import (
	"testing"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
)

type mockRoomRepo struct {
	rooms []domain.Room
}

func (m *mockRoomRepo) Create(room *domain.Room) error {
	m.rooms = append(m.rooms, *room)
	return nil
}

func (m *mockRoomRepo) List() ([]domain.Room, error) {
	return m.rooms, nil
}

func (m *mockRoomRepo) GetByID(id uuid.UUID) (*domain.Room, error) {
	for _, room := range m.rooms {
		if room.ID == id {
			return &room, nil
		}
	}
	return nil, nil
}

func (m *mockRoomRepo) Exists(id uuid.UUID) (bool, error) {
	for _, room := range m.rooms {
		if room.ID == id {
			return true, nil
		}
	}
	return false, nil
}

func TestCreateRoom(t *testing.T) {
	repo := &mockRoomRepo{}
	service := NewRoomService(repo)

	req := &domain.CreateRoomRequest{
		Name:     "Test Room",
		Capacity: intPtr(10),
	}

	room, err := service.CreateRoom(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if room.Name != "Test Room" {
		t.Errorf("Expected name 'Test Room', got %s", room.Name)
	}
}

func TestListRooms(t *testing.T) {
	repo := &mockRoomRepo{}
	service := NewRoomService(repo)

	req := &domain.CreateRoomRequest{Name: "Room 1"}
	service.CreateRoom(req)

	rooms, err := service.ListRooms()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rooms) != 1 {
		t.Errorf("Expected 1 room, got %d", len(rooms))
	}
}

func TestCreateRoom_MultipleRooms(t *testing.T) {
	repo := &mockRoomRepo{}
	service := NewRoomService(repo)

	names := []string{"Room A", "Room B", "Room C"}
	for _, name := range names {
		_, err := service.CreateRoom(&domain.CreateRoomRequest{Name: name})
		if err != nil {
			t.Fatalf("Failed to create room %s: %v", name, err)
		}
	}

	rooms, _ := service.ListRooms()
	if len(rooms) != 3 {
		t.Errorf("Expected 3 rooms, got %d", len(rooms))
	}
}

func TestCreateRoom_EmptyNameError(t *testing.T) {
	repo := &mockRoomRepo{}
	service := NewRoomService(repo)

	_, err := service.CreateRoom(&domain.CreateRoomRequest{Name: ""})
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}
}

func intPtr(i int) *int {
	return &i
}
