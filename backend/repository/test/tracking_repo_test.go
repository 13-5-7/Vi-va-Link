package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bus-logistics/backend/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestTrackingRepo_InsertStatusLog_Success(t *testing.T) {
	bookingID := uuid.New()
	operatorID := uuid.New()
	var capturedOld, capturedNew model.BookingStatus

	repo := &MockTrackingRepo{
		InsertStatusLogFunc: func(_ context.Context, _ pgx.Tx, bID uuid.UUID, old, new model.BookingStatus, changedBy uuid.UUID) error {
			capturedOld = old
			capturedNew = new
			if bID != bookingID {
				t.Errorf("unexpected bookingID: %v", bID)
			}
			if changedBy != operatorID {
				t.Errorf("unexpected changedBy: %v", changedBy)
			}
			return nil
		},
	}

	err := repo.InsertStatusLog(context.Background(), nil, bookingID, model.BookingStatusAccepted, model.BookingStatusLoaded, operatorID)
	if err != nil {
		t.Fatal(err)
	}
	if capturedOld != model.BookingStatusAccepted {
		t.Errorf("want accepted, got %v", capturedOld)
	}
	if capturedNew != model.BookingStatusLoaded {
		t.Errorf("want loaded, got %v", capturedNew)
	}
}

func TestTrackingRepo_InsertStatusLog_Error(t *testing.T) {
	repo := &MockTrackingRepo{
		InsertStatusLogFunc: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _, _ model.BookingStatus, _ uuid.UUID) error {
			return errors.New("db error")
		},
	}

	err := repo.InsertStatusLog(context.Background(), nil, uuid.New(), model.BookingStatusLoaded, model.BookingStatusInTransit, uuid.New())
	if err == nil {
		t.Error("want error, got nil")
	}
}

func TestTrackingRepo_InsertStatusLog_AllTransitions(t *testing.T) {
	transitions := []struct {
		old model.BookingStatus
		new model.BookingStatus
	}{
		{model.BookingStatusAccepted, model.BookingStatusLoaded},
		{model.BookingStatusLoaded, model.BookingStatusInTransit},
		{model.BookingStatusInTransit, model.BookingStatusDelivered},
	}

	for _, tt := range transitions {
		tt := tt
		t.Run(string(tt.old)+"->"+string(tt.new), func(t *testing.T) {
			var capturedOld, capturedNew model.BookingStatus
			repo := &MockTrackingRepo{
				InsertStatusLogFunc: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, old, new model.BookingStatus, _ uuid.UUID) error {
					capturedOld = old
					capturedNew = new
					return nil
				},
			}

			err := repo.InsertStatusLog(context.Background(), nil, uuid.New(), tt.old, tt.new, uuid.New())
			if err != nil {
				t.Fatal(err)
			}
			if capturedOld != tt.old {
				t.Errorf("want old=%v, got %v", tt.old, capturedOld)
			}
			if capturedNew != tt.new {
				t.Errorf("want new=%v, got %v", tt.new, capturedNew)
			}
		})
	}
}
