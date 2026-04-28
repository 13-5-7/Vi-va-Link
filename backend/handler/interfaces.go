package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
)

// AuthServiceInterface はAuthHandlerが依存するサービスのインターフェース
type AuthServiceInterface interface {
	Register(ctx context.Context, req service.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error)
}

// BookingServiceInterface はBookingHandlerが依存するサービスのインターフェース
type BookingServiceInterface interface {
	Create(ctx context.Context, req service.CreateBookingRequest) (*model.Booking, error)
	ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	Cancel(ctx context.Context, bookingID uuid.UUID, shipperID uuid.UUID) error
}

// TrackingServiceInterface はTrackingHandlerが依存するサービスのインターフェース
type TrackingServiceInterface interface {
	GetByTrackingNumber(ctx context.Context, trackingNumber string) (*service.TrackingInfo, error)
	UpdateStatus(ctx context.Context, bookingID uuid.UUID, newStatus model.BookingStatus, operatorID uuid.UUID) error
}

// ScheduleServiceInterface はScheduleHandlerが依存するサービスのインターフェース
type ScheduleServiceInterface interface {
	Create(ctx context.Context, req service.CreateScheduleRequest) (*model.Schedule, error)
	ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error)
	Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
	UpdateScheduleStatus(ctx context.Context, scheduleID uuid.UUID, newStatus model.ScheduleStatus, operatorID uuid.UUID) error
	Delete(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error
	Cancel(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error
}

// scheduleToMap はScheduleをmap[string]anyに変換するヘルパー
func scheduleToMap(s model.Schedule, includeBookings bool) map[string]any {
	log.Println("----interface scheduleToMap called-----")

	m := map[string]any{
		"id":              s.ID,
		"operator_id":     s.OperatorID,
		"origin_lat":      s.OriginLat,
		"origin_lng":      s.OriginLng,
		"origin_name":     s.OriginName,
		"dest_lat":        s.DestLat,
		"dest_lng":        s.DestLng,
		"dest_name":       s.DestName,
		"depart_at":       s.DepartAt,
		"arrive_at":       s.ArriveAt,
		"max_weight_kg":   s.MaxWeightKg,
		"max_size_cm":     s.MaxSizeCm,
		"avail_weight_kg": s.AvailWeightKg,
		"status":          s.Status,
		"route_geojson":   json.RawMessage(s.RouteGeoJSON),
		"created_at":      s.CreatedAt,
	}
	// Bookingsを含めるかどうかはincludeBookingsフラグで制御
	if includeBookings {
		m["bookings"] = s.Bookings
	}
	return m
}
