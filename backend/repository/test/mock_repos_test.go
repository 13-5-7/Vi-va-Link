package repository_test

import (
	"context"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// --- UserRepository インターフェース ---

type UserRepo interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, email, passwordHash string, role model.Role) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type MockUserRepo struct {
	FindByEmailFunc func(ctx context.Context, email string) (*model.User, error)
	CreateFunc      func(ctx context.Context, email, passwordHash string, role model.Role) (*model.User, error)
	FindByIDFunc    func(ctx context.Context, id uuid.UUID) (*model.User, error)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.FindByEmailFunc(ctx, email)
}
func (m *MockUserRepo) Create(ctx context.Context, email, passwordHash string, role model.Role) (*model.User, error) {
	return m.CreateFunc(ctx, email, passwordHash, role)
}
func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return m.FindByIDFunc(ctx, id)
}

// --- BookingRepository インターフェース ---

type BookingRepo interface {
	Create(ctx context.Context, tx pgx.Tx, booking *model.Booking) (*model.Booking, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	FindByTrackingNumber(ctx context.Context, trackingNumber string) (*model.Booking, error)
	ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error)
	ListBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]model.Booking, error)
	UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status model.BookingStatus) error
	UpdateStatusDirect(ctx context.Context, id uuid.UUID, status model.BookingStatus) error
}

type MockBookingRepo struct {
	CreateFunc              func(ctx context.Context, tx pgx.Tx, booking *model.Booking) (*model.Booking, error)
	FindByIDFunc            func(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	FindByTrackingNumberFunc func(ctx context.Context, trackingNumber string) (*model.Booking, error)
	ListByShipperFunc       func(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error)
	ListByScheduleFunc      func(ctx context.Context, scheduleID uuid.UUID) ([]model.Booking, error)
	UpdateStatusFunc        func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status model.BookingStatus) error
	UpdateStatusDirectFunc  func(ctx context.Context, id uuid.UUID, status model.BookingStatus) error
}

func (m *MockBookingRepo) Create(ctx context.Context, tx pgx.Tx, booking *model.Booking) (*model.Booking, error) {
	return m.CreateFunc(ctx, tx, booking)
}
func (m *MockBookingRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	return m.FindByIDFunc(ctx, id)
}
func (m *MockBookingRepo) FindByTrackingNumber(ctx context.Context, trackingNumber string) (*model.Booking, error) {
	return m.FindByTrackingNumberFunc(ctx, trackingNumber)
}
func (m *MockBookingRepo) ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error) {
	return m.ListByShipperFunc(ctx, shipperID)
}
func (m *MockBookingRepo) ListBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]model.Booking, error) {
	return m.ListByScheduleFunc(ctx, scheduleID)
}
func (m *MockBookingRepo) UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status model.BookingStatus) error {
	return m.UpdateStatusFunc(ctx, tx, id, status)
}
func (m *MockBookingRepo) UpdateStatusDirect(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	return m.UpdateStatusDirectFunc(ctx, id, status)
}

// --- ScheduleRepository インターフェース ---

type ScheduleRepo interface {
	Create(ctx context.Context, s *model.Schedule) (*model.Schedule, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
	ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error)
	Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error
}

type MockScheduleRepo struct {
	CreateFunc         func(ctx context.Context, s *model.Schedule) (*model.Schedule, error)
	FindByIDFunc       func(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
	ListByOperatorFunc func(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error)
	SearchFunc         func(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error)
	UpdateStatusFunc   func(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error
	DeleteFunc         func(ctx context.Context, id uuid.UUID) error
}

func (m *MockScheduleRepo) Create(ctx context.Context, s *model.Schedule) (*model.Schedule, error) {
	return m.CreateFunc(ctx, s)
}
func (m *MockScheduleRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	return m.FindByIDFunc(ctx, id)
}
func (m *MockScheduleRepo) ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error) {
	return m.ListByOperatorFunc(ctx, operatorID)
}
func (m *MockScheduleRepo) Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error) {
	return m.SearchFunc(ctx, filter)
}
func (m *MockScheduleRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error {
	return m.UpdateStatusFunc(ctx, id, status)
}
func (m *MockScheduleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}

// --- TrackingRepository インターフェース ---

type TrackingRepo interface {
	InsertStatusLog(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID, oldStatus, newStatus model.BookingStatus, changedBy uuid.UUID) error
}

type MockTrackingRepo struct {
	InsertStatusLogFunc func(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID, oldStatus, newStatus model.BookingStatus, changedBy uuid.UUID) error
}

func (m *MockTrackingRepo) InsertStatusLog(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID, oldStatus, newStatus model.BookingStatus, changedBy uuid.UUID) error {
	return m.InsertStatusLogFunc(ctx, tx, bookingID, oldStatus, newStatus, changedBy)
}
