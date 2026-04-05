package service

import (
	"testing"
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/google/uuid"
)

type mockBookingRepo struct {
	bookings []domain.Booking
}

func (m *mockBookingRepo) Create(booking *domain.Booking) error {
	m.bookings = append(m.bookings, *booking)
	return nil
}

func (m *mockBookingRepo) GetByID(id uuid.UUID) (*domain.Booking, error) {
	for i, b := range m.bookings {
		if b.ID == id {
			return &m.bookings[i], nil
		}
	}
	return nil, nil
}

func (m *mockBookingRepo) UpdateStatus(id uuid.UUID, status string) error {
	for i, b := range m.bookings {
		if b.ID == id {
			m.bookings[i].Status = status
			break
		}
	}
	return nil
}

func (m *mockBookingRepo) GetActiveBookingBySlotID(slotID uuid.UUID) (*domain.Booking, error) {
	for _, b := range m.bookings {
		if b.SlotID == slotID && b.Status == "active" {
			return &b, nil
		}
	}
	return nil, nil
}

func (m *mockBookingRepo) ListAll(offset, limit int) ([]domain.Booking, int, error) {
	return m.bookings, len(m.bookings), nil
}

func (m *mockBookingRepo) ListByUserID(userID uuid.UUID, futureOnly bool) ([]domain.Booking, error) {
	var result []domain.Booking
	for _, b := range m.bookings {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *mockBookingRepo) GetBySlotID(slotID uuid.UUID) (*domain.Booking, error) {
	for _, b := range m.bookings {
		if b.SlotID == slotID && b.Status == "active" {
			return &b, nil
		}
	}
	return nil, nil
}

type mockSlotRepo struct {
	slots []domain.Slot
}

func (m *mockSlotRepo) Create(slot *domain.Slot) error {
	m.slots = append(m.slots, *slot)
	return nil
}

func (m *mockSlotRepo) GetByID(id uuid.UUID) (*domain.Slot, error) {
	for _, s := range m.slots {
		if s.ID == id {
			return &s, nil
		}
	}
	return &domain.Slot{
		ID:     id,
		RoomID: uuid.New(),
		Start:  time.Now().Add(1 * time.Hour),
		End:    time.Now().Add(1*time.Hour + 30*time.Minute),
	}, nil
}

func (m *mockSlotRepo) GetAvailableSlots(roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	return nil, nil
}

func (m *mockSlotRepo) GetBookedSlotIDs(slotIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return make(map[uuid.UUID]bool), nil
}

func (m *mockSlotRepo) GetSlotsForDate(roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	return nil, nil
}

func (m *mockSlotRepo) DeleteSlotsForRoom(roomID uuid.UUID) error {
	return nil
}

type mockUserRepo struct{}

func (m *mockUserRepo) GetByID(id uuid.UUID) (*domain.User, error) {
	return &domain.User{ID: id, Role: "user"}, nil
}

func (m *mockUserRepo) CreateOrGetByID(id uuid.UUID, role string) (*domain.User, error) {
	return &domain.User{ID: id, Role: role}, nil
}

func TestCancelBooking(t *testing.T) {
	bookingRepo := &mockBookingRepo{}
	slotRepo := &mockSlotRepo{}
	userRepo := &mockUserRepo{}
	service := NewBookingService(bookingRepo, slotRepo, userRepo)

	userID := uuid.New()
	slotID := uuid.New()

	slotRepo.Create(&domain.Slot{
		ID:    slotID,
		Start: time.Now().Add(1 * time.Hour),
		End:   time.Now().Add(1*time.Hour + 30*time.Minute),
	})

	req := &domain.CreateBookingRequest{
		SlotID: slotID.String(),
	}
	booking, err := service.CreateBooking(userID, req)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	cancelled, err := service.CancelBooking(booking.ID, userID)
	if err != nil {
		t.Fatalf("Failed to cancel booking: %v", err)
	}

	if cancelled.Status != "cancelled" {
		t.Errorf("Expected status 'cancelled', got '%s'", cancelled.Status)
	}
}

func TestCreateBooking(t *testing.T) {
	bookingRepo := &mockBookingRepo{}
	slotRepo := &mockSlotRepo{}
	userRepo := &mockUserRepo{}
	service := NewBookingService(bookingRepo, slotRepo, userRepo)

	userID := uuid.New()
	slotID := uuid.New()

	slotRepo.Create(&domain.Slot{
		ID:    slotID,
		Start: time.Now().Add(2 * time.Hour),
		End:   time.Now().Add(2*time.Hour + 30*time.Minute),
	})

	req := &domain.CreateBookingRequest{
		SlotID: slotID.String(),
	}

	booking, err := service.CreateBooking(userID, req)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}

	if booking.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", booking.Status)
	}

	if booking.UserID != userID {
		t.Errorf("Expected userID %v, got %v", userID, booking.UserID)
	}
}
