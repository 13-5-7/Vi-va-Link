package handler

import (
	"errors"
	"net/http"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/service"
	"github.com/bus-logistics/backend/utils"
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
	log.Println("----handler GetByTrackingNumber called-----")

	trackingNumber := c.Param("tracking_number")
	if utils.IsEmpty(trackingNumber) {
		return utils.NewAppError(http.StatusBadRequest, "BAD_REQUEST", "tracking_number is required")
	}

	info, err := h.trackingService.GetByTrackingNumber(c.Request().Context(), trackingNumber)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			return utils.NewAppError(http.StatusNotFound, "NOT_FOUND", "tracking number not found")
		default:
			return utils.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":                info.Booking.ID,
		"tracking_number":   info.Booking.TrackingNumber,
		"status":            info.Booking.Status,
		"status_updated_at": info.Booking.StatusUpdatedAt,
		"schedule": map[string]any{
			"origin_name": info.Schedule.OriginName,
			"dest_name":   info.Schedule.DestName,
			"depart_at":   info.Schedule.DepartAt,
		},
	})
}

type updateStatusRequest struct { //nolint:unused
	Status model.BookingStatus `json:"status"`
}

// UpdateStatus handles PATCH /api/v1/bookings/:id/status (JWT: Operator)
func (h *TrackingHandler) UpdateStatus(c echo.Context) error {
	// TODO: ここから自分の手で実装する
    panic("未実装：ここから製造実験開始")
}
