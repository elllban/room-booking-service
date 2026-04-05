package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/elllban/test-backend-elllban/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SlotHandler struct {
	slotService service.SlotService
}

func NewSlotHandler(slotService service.SlotService) *SlotHandler {
	return &SlotHandler{
		slotService: slotService,
	}
}

// ListSlots @Summary Список доступных слотов
// @Description Возвращает список свободных слотов для переговорки на дату
// @Tags Slots
// @Security BearerAuth
// @Param roomId path string true "ID переговорки"
// @Param date query string true "Дата в формате YYYY-MM-DD"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /rooms/{roomId}/slots/list [get]
func (h *SlotHandler) ListSlots(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomIDStr := vars["roomId"]
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		sendError(w, "INVALID_REQUEST", "invalid room id", http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		sendError(w, "INVALID_REQUEST", "date parameter is required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		sendError(w, "INVALID_REQUEST", "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	slots, err := h.slotService.GetAvailableSlots(roomID, date)
	if err != nil {
		sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"slots": slots})
}
