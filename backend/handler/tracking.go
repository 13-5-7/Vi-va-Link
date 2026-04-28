package handler

import (
	//"errors"
	//"net/http"

	"github.com/bus-logistics/backend/model"
	//"github.com/bus-logistics/backend/service"
	//"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TrackingHandler struct {
	trackingService TrackingServiceInterface
}

func NewTrackingHandler(trackingService TrackingServiceInterface) *TrackingHandler {
	return &TrackingHandler{trackingService: trackingService}
}

// GetByTrackingNumber handles GET /api/v1/tracking/:tracking_number (no auth required)
func (h *TrackingHandler) GetByTrackingNumber(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}

type updateStatusRequest struct {
	Status model.BookingStatus `json:"status"`
}

// UpdateStatus handles PATCH /api/v1/bookings/:id/status (JWT: Operator)
func (h *TrackingHandler) UpdateStatus(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
