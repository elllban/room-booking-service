package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/elllban/test-backend-elllban/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// CreateBooking @Summary Создать бронь
// @Description Создаёт бронь на слот (только user)
// @Tags Bookings
// @Security BearerAuth
// @Param request body domain.CreateBookingRequest true "Данные брони"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Router /bookings/create [post]
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok || role != "user" {
		sendError(w, "FORBIDDEN", "only users can create bookings", http.StatusForbidden)
		return
	}

	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		sendError(w, "UNAUTHORIZED", "invalid user context", http.StatusUnauthorized)
		return
	}

	var req domain.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	booking, err := h.bookingService.CreateBooking(userID, &req)
	if err != nil {
		switch err.Error() {
		case "slot not found":
			sendError(w, "SLOT_NOT_FOUND", err.Error(), http.StatusNotFound)
		case "slot is already booked":
			sendError(w, "SLOT_ALREADY_BOOKED", err.Error(), http.StatusConflict)
		case "cannot book a slot in the past":
			sendError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		default:
			sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"booking": booking})
}

// CancelBooking @Summary Отменить бронь
// @Description Отменяет бронь (только свою, только user)
// @Tags Bookings
// @Security BearerAuth
// @Param bookingId path string true "ID брони"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /bookings/{bookingId}/cancel [post]
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok || role != "user" {
		sendError(w, "FORBIDDEN", "only users can cancel bookings", http.StatusForbidden)
		return
	}

	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		sendError(w, "UNAUTHORIZED", "invalid user context", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	bookingIDStr := vars["bookingId"]
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		sendError(w, "INVALID_REQUEST", "invalid booking id", http.StatusBadRequest)
		return
	}

	booking, err := h.bookingService.CancelBooking(bookingID, userID)
	if err != nil {
		if err.Error() == "booking not found" {
			sendError(w, "BOOKING_NOT_FOUND", err.Error(), http.StatusNotFound)
		} else if err.Error() == "cannot cancel another user's booking" {
			sendError(w, "FORBIDDEN", err.Error(), http.StatusForbidden)
		} else {
			sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"booking": booking})
}

// ListAllBookings @Summary Список всех броней
// @Description Возвращает список всех броней с пагинацией (только admin)
// @Tags Bookings
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1)
// @Param pageSize query int false "Размер страницы" default(20) maximum(100)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Router /bookings/list [get]
func (h *BookingHandler) ListAllBookings(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok || role != "admin" {
		sendError(w, "FORBIDDEN", "only admin can list all bookings", http.StatusForbidden)
		return
	}

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 1 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps >= 1 && ps <= 100 {
			pageSize = ps
		}
	}

	bookings, total, err := h.bookingService.ListAllBookings(page, pageSize)
	if err != nil {
		sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"bookings": bookings,
		"pagination": map[string]interface{}{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListMyBookings @Summary Мои брони
// @Description Возвращает список броней текущего пользователя (только user)
// @Tags Bookings
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Router /bookings/my [get]
func (h *BookingHandler) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok || role != "user" {
		sendError(w, "FORBIDDEN", "only users can list their bookings", http.StatusForbidden)
		return
	}

	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		sendError(w, "UNAUTHORIZED", "invalid user context", http.StatusUnauthorized)
		return
	}

	bookings, err := h.bookingService.ListMyBookings(userID)
	if err != nil {
		sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"bookings": bookings})
}
