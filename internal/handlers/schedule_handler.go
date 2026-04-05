package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/elllban/test-backend-elllban/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ScheduleHandler struct {
	scheduleService service.ScheduleService
}

func NewScheduleHandler(scheduleService service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
	}
}

// CreateSchedule @Summary Создать расписание
// @Description Создаёт расписание для переговорки (только admin, один раз)
// @Tags Schedules
// @Security BearerAuth
// @Param roomId path string true "ID переговорки"
// @Param request body domain.CreateScheduleRequest true "Данные расписания"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Router /rooms/{roomId}/schedule/create [post]
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok || role != "admin" {
		sendError(w, "FORBIDDEN", "only admin can create schedules", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	roomIDStr := vars["roomId"]
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		sendError(w, "INVALID_REQUEST", "invalid room id", http.StatusBadRequest)
		return
	}

	var req domain.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	schedule, err := h.scheduleService.CreateSchedule(roomID, &req)
	if err != nil {
		if err.Error() == "room not found" {
			sendError(w, "ROOM_NOT_FOUND", err.Error(), http.StatusNotFound)
		} else if err.Error() == "schedule already exists for this room" {
			sendError(w, "SCHEDULE_EXISTS", err.Error(), http.StatusConflict)
		} else {
			sendError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"schedule": schedule})
}
