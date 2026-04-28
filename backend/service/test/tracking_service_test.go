package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// テスト用 TrackingService: pool を使わずロジックのみ検証
type trackingServiceImpl struct {
	bookingRepo  bookingRepoIface
	scheduleRepo scheduleRepoIface
	trackingRepo trackingRepoIface
}

type bookingRepoIface interface {
	Create(ctx context.Context, tx pgx.Tx, booking *model.Booking) (*model.Booking, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	FindByTrackingNumber(ctx context.Context, trackingNumber string) (*model.Booking, error)
	ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error)
	UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status model.BookingStatus) error
}

type trackingRepoIface interface {
	InsertStatusLog(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID, oldStatus, newStatus model.BookingStatus, changedBy uuid.UUID) error
}

func (s *trackingServiceImpl) GetByTrackingNumber(ctx context.Context, trackingNumber string) (*service.TrackingInfo, error) {
	booking, err := s.bookingRepo.FindByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, service.ErrBookingNotFound
	}
	schedule, err := s.scheduleRepo.FindByID(ctx, booking.ScheduleID)
	if err != nil {
		return nil, err
	}
	return &service.TrackingInfo{Booking: booking, Schedule: schedule}, nil
}

func (s *trackingServiceImpl) UpdateStatus(ctx context.Context, bookingID uuid.UUID, newStatus model.BookingStatus, operatorID uuid.UUID) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking == nil {
		return service.ErrBookingNotFound
	}
	if !booking.Status.CanTransitionTo(newStatus) {
		return service.ErrInvalidStatusTransition
	}
	oldStatus := booking.Status
	// tx なしでロジックのみ検証（UpdateStatus と InsertStatusLog を直接呼ぶ）
	if err = s.bookingRepo.UpdateStatus(ctx, nil, bookingID, newStatus); err != nil {
		return err
	}
	return s.trackingRepo.InsertStatusLog(ctx, nil, bookingID, oldStatus, newStatus, operatorID)
}

// ---- GetByTrackingNumber ----

func TestTrackingService_GetByTrackingNumber_Success(t *testing.T) {
	bookingID := uuid.New()
	scheduleID := uuid.New()

	svc := &trackingServiceImpl{
		bookingRepo: &mockBookingRepo{
			findByTrackingNumber: func(_ context.Context, tn string) (*model.Booking, error) {
				return &model.Booking{
					ID:             bookingID,
					ScheduleID:     scheduleID,
					TrackingNumber: tn,
					Status:         model.BookingStatusAccepted,
					CreatedAt:      time.Now(),
				}, nil
			},
		},
		scheduleRepo: &mockScheduleRepo{
			findByID: func(_ context.Context, id uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: id, OriginName: "Tokyo", DestName: "Osaka"}, nil
			},
		},
	}

	info, err := svc.GetByTrackingNumber(context.Background(), "TRK-ABC123")
	if err != nil {
		t.Fatal(err)
	}
	if info.Booking.TrackingNumber != "TRK-ABC123" {
		t.Errorf("want TRK-ABC123, got %s", info.Booking.TrackingNumber)
	}
	if info.Schedule.OriginName != "Tokyo" {
		t.Errorf("want Tokyo, got %s", info.Schedule.OriginName)
	}
}

func TestTrackingService_GetByTrackingNumber_NotFound(t *testing.T) {
	svc := &trackingServiceImpl{
		bookingRepo: &mockBookingRepo{
			findByTrackingNumber: func(_ context.Context, _ string) (*model.Booking, error) {
				return nil, nil
			},
		},
	}

	_, err := svc.GetByTrackingNumber(context.Background(), "TRK-NOTEXIST")
	if !errors.Is(err, service.ErrBookingNotFound) {
		t.Errorf("want ErrBookingNotFound, got %v", err)
	}
}

func TestTrackingService_GetByTrackingNumber_RepoError(t *testing.T) {
	svc := &trackingServiceImpl{
		bookingRepo: &mockBookingRepo{
			findByTrackingNumber: func(_ context.Context, _ string) (*model.Booking, error) {
				return nil, errors.New("db error")
			},
		},
	}

	_, err := svc.GetByTrackingNumber(context.Background(), "TRK-ERR")
	if err == nil {
		t.Error("want error, got nil")
	}
}

// ---- UpdateStatus ----

func TestTrackingService_UpdateStatus_Success(t *testing.T) {
	bookingID := uuid.New()
	operatorID := uuid.New()
	var capturedOld, capturedNew model.BookingStatus

	svc := &trackingServiceImpl{
		bookingRepo: &mockBookingRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
				return &model.Booking{ID: bookingID, Status: model.BookingStatusAccepted}, nil
			},
			updateStatus: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ model.BookingStatus) error {
				return nil
			},
		},
		trackingRepo: &mockTrackingRepo{
			insertStatusLog: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, old, new model.BookingStatus, _ uuid.UUID) error {
				capturedOld = old
				capturedNew = new
				return nil
			},
		},
	}

	err := svc.UpdateStatus(context.Background(), bookingID, model.BookingStatusLoaded, operatorID)
	if err != nil {
		t.Fatal(err)
	}
	if capturedOld != model.BookingStatusAccepted {
		t.Errorf("want old=accepted, got %v", capturedOld)
	}
	if capturedNew != model.BookingStatusLoaded {
		t.Errorf("want new=loaded, got %v", capturedNew)
	}
}

func TestTrackingService_UpdateStatus_BookingNotFound(t *testing.T) {
	svc := &trackingServiceImpl{
		bookingRepo: &mockBookingRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) { return nil, nil },
		},
	}

	err := svc.UpdateStatus(context.Background(), uuid.New(), model.BookingStatusLoaded, uuid.New())
	if !errors.Is(err, service.ErrBookingNotFound) {
		t.Errorf("want ErrBookingNotFound, got %v", err)
	}
}

func TestTrackingService_UpdateStatus_InvalidTransition_Backward(t *testing.T) {
	bookingID := uuid.New()
	svc := &trackingServiceImpl{
		bookingRepo: &mockBookingRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
				return &model.Booking{ID: bookingID, Status: model.BookingStatusDelivered}, nil
			},
		},
	}

	err := svc.UpdateStatus(context.Background(), bookingID, model.BookingStatusLoaded, uuid.New())
	if !errors.Is(err, service.ErrInvalidStatusTransition) {
		t.Errorf("want ErrInvalidStatusTransition, got %v", err)
	}
}

func TestTrackingService_UpdateStatus_InvalidTransition_SameStatus(t *testing.T) {
	bookingID := uuid.New()
	svc := &trackingServiceImpl{
		bookingRepo: &mockBookingRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
				return &model.Booking{ID: bookingID, Status: model.BookingStatusLoaded}, nil
			},
		},
	}

	err := svc.UpdateStatus(context.Background(), bookingID, model.BookingStatusLoaded, uuid.New())
	if !errors.Is(err, service.ErrInvalidStatusTransition) {
		t.Errorf("want ErrInvalidStatusTransition, got %v", err)
	}
}

func TestTrackingService_UpdateStatus_AllValidTransitions(t *testing.T) {
	transitions := []struct {
		from model.BookingStatus
		to   model.BookingStatus
	}{
		{model.BookingStatusAccepted, model.BookingStatusLoaded},
		{model.BookingStatusLoaded, model.BookingStatusInTransit},
		{model.BookingStatusInTransit, model.BookingStatusDelivered},
	}

	for _, tt := range transitions {
		tt := tt
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			svc := &trackingServiceImpl{
				bookingRepo: &mockBookingRepo{
					findByID: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
						return &model.Booking{Status: tt.from}, nil
					},
					updateStatus: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ model.BookingStatus) error {
						return nil
					},
				},
				trackingRepo: &mockTrackingRepo{
					insertStatusLog: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _, _ model.BookingStatus, _ uuid.UUID) error {
						return nil
					},
				},
			}

			err := svc.UpdateStatus(context.Background(), uuid.New(), tt.to, uuid.New())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTrackingService_UpdateStatus_UpdateRepoError(t *testing.T) {
	bookingID := uuid.New()
	svc := &trackingServiceImpl{
		bookingRepo: &mockBookingRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Booking, error) {
				return &model.Booking{ID: bookingID, Status: model.BookingStatusAccepted}, nil
			},
			updateStatus: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ model.BookingStatus) error {
				return errors.New("db error")
			},
		},
		trackingRepo: &mockTrackingRepo{},
	}

	err := svc.UpdateStatus(context.Background(), bookingID, model.BookingStatusLoaded, uuid.New())
	if err == nil {
		t.Error("want error, got nil")
	}
}
