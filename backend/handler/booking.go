package handler

import (
	//"errors"
	//"net/http"

	//"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BookingHandler struct {
	bookingService BookingServiceInterface
}

func NewBookingHandler(bookingService BookingServiceInterface) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

type createBookingRequest struct {
	ScheduleID     uuid.UUID `json:"schedule_id"`
	WeightKg       float64   `json:"weight_kg"`
	SizeCm         float64   `json:"size_cm"`
	ContentDesc    string    `json:"content_desc"`
	RecipientName  string    `json:"recipient_name"`
	RecipientPhone string    `json:"recipient_phone"`
	RecipientAddr  string    `json:"recipient_addr"`
}

// List returns the shipper's own bookings
func (h *BookingHandler) List(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// Create creates a new booking for the shipper
func (h *BookingHandler) Create(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// GetByID returns a single booking by ID
func (h *BookingHandler) GetByID(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

// Cancel cancels a booking (only if status is accepted)
func (h *BookingHandler) Cancel(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
