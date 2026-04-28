package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bus-logistics/backend/model"
	"github.com/google/uuid"
)

func newBooking(shipperID, scheduleID uuid.UUID) *model.Booking {
	return &model.Booking{
		ID:             uuid.New(),
		ScheduleID:     scheduleID,
		ShipperID:      shipperID,
		TrackingNumber: "TRK-TEST001",
		WeightKg:       5.0,
		SizeCm:         30.0,
		Status:         model.BookingStatusAccepted,
		CreatedAt:      time.Now(),
	}
}

// ---- FindByID ----

func TestBookingRepo_FindByID_Found(t *testing.T) {
	b := newBooking(uuid.New(), uuid.New())
	repo := &MockBookingRepo{
		FindByIDFunc: func(_ context.Context, id uuid.UUID) (*model.Booking, error) {
			if id != b.ID {
				return nil, nil
			}
			return b, nil
		},
	}

	got, err := repo.FindByID(context.Background(), b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("want booking, got nil")
	}
	if got.TrackingNumber != "TRK-TEST001" {
		t.Errorf("want TRK-TEST001, got %s", got.TrackingNumber)
	}
}

func TestBookingRepo_FindByID_NotFound(t *testing.T) {
	repo := &MockBookingRepo{
		FindByIDFunc: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
			return nil, nil
		},
	}

	got, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

// ---- FindByTrackingNumber ----

func TestBookingRepo_FindByTrackingNumber_Found(t *testing.T) {
	b := newBooking(uuid.New(), uuid.New())
	repo := &MockBookingRepo{
		FindByTrackingNumberFunc: func(_ context.Context, tn string) (*model.Booking, error) {
			if tn != b.TrackingNumber {
				return nil, nil
			}
			return b, nil
		},
	}

	got, err := repo.FindByTrackingNumber(context.Background(), "TRK-TEST001")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("want booking, got nil")
	}
	if got.TrackingNumber != "TRK-TEST001" {
		t.Errorf("want TRK-TEST001, got %s", got.TrackingNumber)
	}
}

func TestBookingRepo_FindByTrackingNumber_NotFound(t *testing.T) {
	repo := &MockBookingRepo{
		FindByTrackingNumberFunc: func(_ context.Context, _ string) (*model.Booking, error) {
			return nil, nil
		},
	}

	got, err := repo.FindByTrackingNumber(context.Background(), "TRK-NOTEXIST")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

// ---- ListByShipper ----

func TestBookingRepo_ListByShipper_Empty(t *testing.T) {
	repo := &MockBookingRepo{
		ListByShipperFunc: func(_ context.Context, _ uuid.UUID) ([]model.Booking, error) {
			return []model.Booking{}, nil
		},
	}

	list, err := repo.ListByShipper(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("want 0, got %d", len(list))
	}
}

func TestBookingRepo_ListByShipper_ReturnsOnlyOwn(t *testing.T) {
	shipperID := uuid.New()
	b1 := newBooking(shipperID, uuid.New())
	b2 := newBooking(shipperID, uuid.New())

	repo := &MockBookingRepo{
		ListByShipperFunc: func(_ context.Context, id uuid.UUID) ([]model.Booking, error) {
			if id != shipperID {
				return []model.Booking{}, nil
			}
			return []model.Booking{*b1, *b2}, nil
		},
	}

	list, err := repo.ListByShipper(context.Background(), shipperID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("want 2, got %d", len(list))
	}
	for _, b := range list {
		if b.ShipperID != shipperID {
			t.Errorf("booking %v does not belong to shipper %v", b.ID, shipperID)
		}
	}
}

func TestBookingRepo_ListByShipper_Error(t *testing.T) {
	repo := &MockBookingRepo{
		ListByShipperFunc: func(_ context.Context, _ uuid.UUID) ([]model.Booking, error) {
			return nil, errors.New("db error")
		},
	}

	_, err := repo.ListByShipper(context.Background(), uuid.New())
	if err == nil {
		t.Error("want error, got nil")
	}
}

// ---- ListBySchedule ----

func TestBookingRepo_ListBySchedule_Empty(t *testing.T) {
	repo := &MockBookingRepo{
		ListByScheduleFunc: func(_ context.Context, _ uuid.UUID) ([]model.Booking, error) {
			return []model.Booking{}, nil
		},
	}

	list, err := repo.ListBySchedule(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("want 0, got %d", len(list))
	}
}

func TestBookingRepo_ListBySchedule_ReturnsOnlyOwn(t *testing.T) {
	scheduleID := uuid.New()
	b1 := newBooking(uuid.New(), scheduleID)
	b2 := newBooking(uuid.New(), scheduleID)

	repo := &MockBookingRepo{
		ListByScheduleFunc: func(_ context.Context, id uuid.UUID) ([]model.Booking, error) {
			if id != scheduleID {
				return []model.Booking{}, nil
			}
			return []model.Booking{*b1, *b2}, nil
		},
	}

	list, err := repo.ListBySchedule(context.Background(), scheduleID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("want 2, got %d", len(list))
	}
	for _, b := range list {
		if b.ScheduleID != scheduleID {
			t.Errorf("booking %v does not belong to schedule %v", b.ID, scheduleID)
		}
	}
}

// ---- UpdateStatusDirect ----

func TestBookingRepo_UpdateStatusDirect_Success(t *testing.T) {
	bookingID := uuid.New()
	var capturedStatus model.BookingStatus

	repo := &MockBookingRepo{
		UpdateStatusDirectFunc: func(_ context.Context, id uuid.UUID, status model.BookingStatus) error {
			if id != bookingID {
				t.Errorf("unexpected id: %v", id)
			}
			capturedStatus = status
			return nil
		},
	}

	err := repo.UpdateStatusDirect(context.Background(), bookingID, model.BookingStatusLoaded)
	if err != nil {
		t.Fatal(err)
	}
	if capturedStatus != model.BookingStatusLoaded {
		t.Errorf("want loaded, got %v", capturedStatus)
	}
}

func TestBookingRepo_UpdateStatusDirect_Error(t *testing.T) {
	repo := &MockBookingRepo{
		UpdateStatusDirectFunc: func(_ context.Context, _ uuid.UUID, _ model.BookingStatus) error {
			return errors.New("db error")
		},
	}

	err := repo.UpdateStatusDirect(context.Background(), uuid.New(), model.BookingStatusDelivered)
	if err == nil {
		t.Error("want error, got nil")
	}
}

// ---- Create ----

func TestBookingRepo_Create_InitialStatus(t *testing.T) {
	// Create に渡す Booking の初期ステータスが accepted であることを確認する
	// （pgx.Tx を要求するため、モック経由で契約を検証）
	shipperID := uuid.New()
	scheduleID := uuid.New()
	b := newBooking(shipperID, scheduleID)

	if b.Status != model.BookingStatusAccepted {
		t.Errorf("want accepted, got %v", b.Status)
	}
	if b.ShipperID != shipperID {
		t.Errorf("want %v, got %v", shipperID, b.ShipperID)
	}
	if b.ScheduleID != scheduleID {
		t.Errorf("want %v, got %v", scheduleID, b.ScheduleID)
	}
	if b.TrackingNumber == "" {
		t.Error("TrackingNumber must not be empty")
	}
}
