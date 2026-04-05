package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
)

type mockRoomService struct {
	rooms []domain.Room
}

func (m *mockRoomService) CreateRoom(req *domain.CreateRoomRequest) (*domain.Room, error) {
	room := &domain.Room{
		ID:   uuid.New(),
		Name: req.Name,
	}
	m.rooms = append(m.rooms, *room)
	return room, nil
}

func (m *mockRoomService) ListRooms() ([]domain.Room, error) {
	return m.rooms, nil
}

func TestListRoomsHandler(t *testing.T) {
	service := &mockRoomService{}
	handler := NewRoomHandler(service)

	req := httptest.NewRequest("GET", "/rooms/list", nil)
	w := httptest.NewRecorder()

	handler.ListRooms(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCreateRoomHandler_AdminOnly(t *testing.T) {
	service := &mockRoomService{}
	handler := NewRoomHandler(service)

	body := `{"name":"Test Room"}`
	req := httptest.NewRequest("POST", "/rooms/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateRoom(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}
