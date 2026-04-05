package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/elllban/test-backend-elllban/internal/service"
)

type RoomHandler struct {
	roomService service.RoomService
}

func NewRoomHandler(roomService service.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

// ListRooms @Summary Список переговорок
// @Description Возвращает список всех переговорок
// @Tags Rooms
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /rooms/list [get]
func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomService.ListRooms()
	if err != nil {
		sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"rooms": rooms})
}

// CreateRoom @Summary Создать переговорку
// @Description Создаёт новую переговорку (только admin)
// @Tags Rooms
// @Security BearerAuth
// @Param request body domain.CreateRoomRequest true "Данные переговорки"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /rooms/create [post]
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok || role != "admin" {
		sendError(w, "FORBIDDEN", "only admin can create rooms", http.StatusForbidden)
		return
	}

	var req domain.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	room, err := h.roomService.CreateRoom(&req)
	if err != nil {
		sendError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"room": room})
}
