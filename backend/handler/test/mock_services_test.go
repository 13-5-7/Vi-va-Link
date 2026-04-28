package handler_test

import (
	"context"
	"encoding/json"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
)

// --- AuthService モック ---

type MockAuthService struct {
	RegisterFunc func(ctx context.Context, req service.RegisterRequest) (*model.User, error)
	LoginFunc    func(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error)
}

func (m *MockAuthService) Register(ctx context.Context, req service.RegisterRequest) (*model.User, error) {
	return m.RegisterFunc(ctx, req)
}

func (m *MockAuthService) Login(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error) {
	return m.LoginFunc(ctx, req)
}

// --- BookingService モック ---

type MockBookingService struct {
	CreateFunc        func(ctx context.Context, req service.CreateBookingRequest) (*model.Booking, error)
	ListByShipperFunc func(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error)
	GetByIDFunc       func(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	CancelFunc        func(ctx context.Context, bookingID uuid.UUID, shipperID uuid.UUID) error
}

func (m *MockBookingService) Create(ctx context.Context, req service.CreateBookingRequest) (*model.Booking, error) {
	return m.CreateFunc(ctx, req)
}

func (m *MockBookingService) ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error) {
	return m.ListByShipperFunc(ctx, shipperID)
}

func (m *MockBookingService) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockBookingService) Cancel(ctx context.Context, bookingID uuid.UUID, shipperID uuid.UUID) error {
	return m.CancelFunc(ctx, bookingID, shipperID)
}

// --- TrackingService モック ---

type MockTrackingService struct {
	GetByTrackingNumberFunc func(ctx context.Context, trackingNumber string) (*service.TrackingInfo, error)
	UpdateStatusFunc        func(ctx context.Context, bookingID uuid.UUID, newStatus model.BookingStatus, operatorID uuid.UUID) error
}

func (m *MockTrackingService) GetByTrackingNumber(ctx context.Context, trackingNumber string) (*service.TrackingInfo, error) {
	return m.GetByTrackingNumberFunc(ctx, trackingNumber)
}

func (m *MockTrackingService) UpdateStatus(ctx context.Context, bookingID uuid.UUID, newStatus model.BookingStatus, operatorID uuid.UUID) error {
	return m.UpdateStatusFunc(ctx, bookingID, newStatus, operatorID)
}

// --- ScheduleService モック ---

type MockScheduleService struct {
	CreateFunc               func(ctx context.Context, req service.CreateScheduleRequest) (*model.Schedule, error)
	ListByOperatorFunc       func(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error)
	SearchFunc               func(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error)
	GetByIDFunc              func(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
	UpdateScheduleStatusFunc func(ctx context.Context, scheduleID uuid.UUID, newStatus model.ScheduleStatus, operatorID uuid.UUID) error
	DeleteFunc               func(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error
	CancelFunc               func(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error
}

func (m *MockScheduleService) Create(ctx context.Context, req service.CreateScheduleRequest) (*model.Schedule, error) {
	return m.CreateFunc(ctx, req)
}

func (m *MockScheduleService) ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error) {
	return m.ListByOperatorFunc(ctx, operatorID)
}

func (m *MockScheduleService) Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error) {
	return m.SearchFunc(ctx, filter)
}

func (m *MockScheduleService) GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockScheduleService) UpdateScheduleStatus(ctx context.Context, scheduleID uuid.UUID, newStatus model.ScheduleStatus, operatorID uuid.UUID) error {
	return m.UpdateScheduleStatusFunc(ctx, scheduleID, newStatus, operatorID)
}

func (m *MockScheduleService) Delete(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error {
	return m.DeleteFunc(ctx, scheduleID, operatorID)
}

func (m *MockScheduleService) Cancel(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error {
	return m.CancelFunc(ctx, scheduleID, operatorID)
}

// コンパイルエラー回避
var _ = json.Marshal
