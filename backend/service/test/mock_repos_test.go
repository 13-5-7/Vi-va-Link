package service_test

import (
	"context"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// --- userRepo モック ---

type mockUserRepo struct {
	findByEmail func(ctx context.Context, email string) (*model.User, error)
	create      func(ctx context.Context, email, hash string, role model.Role) (*model.User, error)
	findByID    func(ctx context.Context, id uuid.UUID) (*model.User, error)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.findByEmail(ctx, email)
}
func (m *mockUserRepo) Create(ctx context.Context, email, hash string, role model.Role) (*model.User, error) {
	return m.create(ctx, email, hash, role)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return m.findByID(ctx, id)
}

// --- scheduleRepo モック ---

type mockScheduleRepo struct {
	create         func(ctx context.Context, s *model.Schedule) (*model.Schedule, error)
	findByID       func(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
	listByOperator func(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error)
	search         func(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error)
	updateStatus   func(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error
	delete         func(ctx context.Context, id uuid.UUID) error
}

func (m *mockScheduleRepo) Create(ctx context.Context, s *model.Schedule) (*model.Schedule, error) {
	return m.create(ctx, s)
}
func (m *mockScheduleRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	return m.findByID(ctx, id)
}
func (m *mockScheduleRepo) ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error) {
	return m.listByOperator(ctx, operatorID)
}
func (m *mockScheduleRepo) Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error) {
	return m.search(ctx, filter)
}
func (m *mockScheduleRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error {
	return m.updateStatus(ctx, id, status)
}
func (m *mockScheduleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.delete(ctx, id)
}

// --- bookingRepo モック ---

type mockBookingRepo struct {
	create               func(ctx context.Context, tx pgx.Tx, booking *model.Booking) (*model.Booking, error)
	findByID             func(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	findByTrackingNumber func(ctx context.Context, trackingNumber string) (*model.Booking, error)
	listByShipper        func(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error)
	updateStatus         func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status model.BookingStatus) error
}

func (m *mockBookingRepo) Create(ctx context.Context, tx pgx.Tx, booking *model.Booking) (*model.Booking, error) {
	return m.create(ctx, tx, booking)
}
func (m *mockBookingRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	return m.findByID(ctx, id)
}
func (m *mockBookingRepo) FindByTrackingNumber(ctx context.Context, tn string) (*model.Booking, error) {
	return m.findByTrackingNumber(ctx, tn)
}
func (m *mockBookingRepo) ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error) {
	return m.listByShipper(ctx, shipperID)
}
func (m *mockBookingRepo) UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status model.BookingStatus) error {
	return m.updateStatus(ctx, tx, id, status)
}

// --- trackingRepo モック ---

type mockTrackingRepo struct {
	insertStatusLog func(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID, oldStatus, newStatus model.BookingStatus, changedBy uuid.UUID) error
}

func (m *mockTrackingRepo) InsertStatusLog(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID, oldStatus, newStatus model.BookingStatus, changedBy uuid.UUID) error {
	return m.insertStatusLog(ctx, tx, bookingID, oldStatus, newStatus, changedBy)
}
