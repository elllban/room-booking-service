package service

import (
	"errors"
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/elllban/test-backend-elllban/internal/repository"
	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(userID uuid.UUID, req *domain.CreateBookingRequest) (*domain.Booking, error)
	CancelBooking(bookingID, userID uuid.UUID) (*domain.Booking, error)
	ListAllBookings(page, pageSize int) ([]domain.Booking, int, error)
	ListMyBookings(userID uuid.UUID) ([]domain.Booking, error)
}

type bookingService struct {
	bookingRepo repository.BookingRepository
	slotRepo    repository.SlotRepository
	userRepo    repository.UserRepository
}

func NewBookingService(bookingRepo repository.BookingRepository, slotRepo repository.SlotRepository, userRepo repository.UserRepository) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		slotRepo:    slotRepo,
		userRepo:    userRepo,
	}
}

func (s *bookingService) CreateBooking(userID uuid.UUID, req *domain.CreateBookingRequest) (*domain.Booking, error) {
	slotID, err := uuid.Parse(req.SlotID)
	if err != nil {
		return nil, errors.New("invalid slot id")
	}

	slot, err := s.slotRepo.GetByID(slotID)
	if err != nil {
		return nil, err
	}
	if slot == nil {
		return nil, errors.New("slot not found")
	}

	if slot.Start.Before(time.Now()) {
		return nil, errors.New("cannot book a slot in the past")
	}

	existingBooking, err := s.bookingRepo.GetActiveBookingBySlotID(slotID)
	if err != nil {
		return nil, err
	}
	if existingBooking != nil {
		return nil, errors.New("slot is already booked")
	}

	var conferenceLink *string
	if req.CreateConferenceLink {
		link := "https://meet.example.com/" + uuid.New().String()
		conferenceLink = &link
	}

	booking := &domain.Booking{
		ID:             uuid.New(),
		SlotID:         slotID,
		UserID:         userID,
		Status:         "active",
		ConferenceLink: conferenceLink,
	}

	if err = s.bookingRepo.Create(booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *bookingService) CancelBooking(bookingID, userID uuid.UUID) (*domain.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, errors.New("booking not found")
	}

	if booking.UserID != userID {
		return nil, errors.New("cannot cancel another user's booking")
	}

	if booking.Status != "cancelled" {
		if err = s.bookingRepo.UpdateStatus(bookingID, "cancelled"); err != nil {
			return nil, err
		}
		booking.Status = "cancelled"
	}

	return booking, nil
}

func (s *bookingService) ListAllBookings(page, pageSize int) ([]domain.Booking, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	return s.bookingRepo.ListAll(offset, pageSize)
}

func (s *bookingService) ListMyBookings(userID uuid.UUID) ([]domain.Booking, error) {
	return s.bookingRepo.ListByUserID(userID, true)
}
