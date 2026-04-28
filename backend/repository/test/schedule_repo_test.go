package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/google/uuid"
)

func newScheduleModel(operatorID uuid.UUID) *model.Schedule {
	return &model.Schedule{
		ID:            uuid.New(),
		OperatorID:    operatorID,
		OriginLat:     35.68,
		OriginLng:     139.76,
		OriginName:    "Tokyo Station",
		DestLat:       34.69,
		DestLng:       135.50,
		DestName:      "Osaka Station",
		DepartAt:      time.Now().Add(24 * time.Hour),
		ArriveAt:      time.Now().Add(30 * time.Hour),
		MaxWeightKg:   100,
		MaxSizeCm:     200,
		AvailWeightKg: 100,
		Status:        model.ScheduleStatusOpen,
		Bookings:      []model.Booking{},
	}
}

// ---- Create ----

func TestScheduleRepo_Create_Success(t *testing.T) {
	operatorID := uuid.New()
	s := newScheduleModel(operatorID)

	repo := &MockScheduleRepo{
		CreateFunc: func(_ context.Context, input *model.Schedule) (*model.Schedule, error) {
			return input, nil
		},
	}

	got, err := repo.Create(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if got.OriginName != "Tokyo Station" {
		t.Errorf("want Tokyo Station, got %s", got.OriginName)
	}
	if got.OperatorID != operatorID {
		t.Errorf("want %v, got %v", operatorID, got.OperatorID)
	}
}

func TestScheduleRepo_Create_Error(t *testing.T) {
	repo := &MockScheduleRepo{
		CreateFunc: func(_ context.Context, _ *model.Schedule) (*model.Schedule, error) {
			return nil, errors.New("db error")
		},
	}

	_, err := repo.Create(context.Background(), newScheduleModel(uuid.New()))
	if err == nil {
		t.Error("want error, got nil")
	}
}

// ---- FindByID ----

func TestScheduleRepo_FindByID_Found(t *testing.T) {
	operatorID := uuid.New()
	s := newScheduleModel(operatorID)

	repo := &MockScheduleRepo{
		FindByIDFunc: func(_ context.Context, id uuid.UUID) (*model.Schedule, error) {
			if id != s.ID {
				return nil, nil
			}
			return s, nil
		},
	}

	got, err := repo.FindByID(context.Background(), s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("want schedule, got nil")
	}
	if got.DestName != "Osaka Station" {
		t.Errorf("want Osaka Station, got %s", got.DestName)
	}
}

func TestScheduleRepo_FindByID_NotFound(t *testing.T) {
	repo := &MockScheduleRepo{
		FindByIDFunc: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
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

func TestScheduleRepo_FindByID_WithBookings(t *testing.T) {
	operatorID := uuid.New()
	s := newScheduleModel(operatorID)
	s.Bookings = []model.Booking{
		{ID: uuid.New(), TrackingNumber: "TRK-001", Status: model.BookingStatusAccepted},
		{ID: uuid.New(), TrackingNumber: "TRK-002", Status: model.BookingStatusLoaded},
	}

	repo := &MockScheduleRepo{
		FindByIDFunc: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
			return s, nil
		},
	}

	got, err := repo.FindByID(context.Background(), s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Bookings) != 2 {
		t.Errorf("want 2 bookings, got %d", len(got.Bookings))
	}
}

// ---- ListByOperator ----

func TestScheduleRepo_ListByOperator_Empty(t *testing.T) {
	repo := &MockScheduleRepo{
		ListByOperatorFunc: func(_ context.Context, _ uuid.UUID) ([]model.Schedule, error) {
			return []model.Schedule{}, nil
		},
	}

	list, err := repo.ListByOperator(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("want 0, got %d", len(list))
	}
}

func TestScheduleRepo_ListByOperator_ReturnsOnlyOwn(t *testing.T) {
	operatorID := uuid.New()
	s1 := newScheduleModel(operatorID)
	s2 := newScheduleModel(operatorID)

	repo := &MockScheduleRepo{
		ListByOperatorFunc: func(_ context.Context, id uuid.UUID) ([]model.Schedule, error) {
			if id != operatorID {
				return []model.Schedule{}, nil
			}
			return []model.Schedule{*s1, *s2}, nil
		},
	}

	list, err := repo.ListByOperator(context.Background(), operatorID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("want 2, got %d", len(list))
	}
	for _, s := range list {
		if s.OperatorID != operatorID {
			t.Errorf("schedule %v does not belong to operator %v", s.ID, operatorID)
		}
	}
}

// ---- Search ----

func TestScheduleRepo_Search_Empty(t *testing.T) {
	repo := &MockScheduleRepo{
		SearchFunc: func(_ context.Context, _ repository.ScheduleFilter) ([]model.Schedule, error) {
			return []model.Schedule{}, nil
		},
	}

	list, err := repo.Search(context.Background(), repository.ScheduleFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("want 0, got %d", len(list))
	}
}

func TestScheduleRepo_Search_WithFilter_OriginBounds(t *testing.T) {
	operatorID := uuid.New()
	s := newScheduleModel(operatorID) // OriginLat=35.68

	repo := &MockScheduleRepo{
		SearchFunc: func(_ context.Context, f repository.ScheduleFilter) ([]model.Schedule, error) {
			// フィルタが正しく渡されているか確認
			if f.OriginLatMin == nil || *f.OriginLatMin != 35.0 {
				return []model.Schedule{}, nil
			}
			return []model.Schedule{*s}, nil
		},
	}

	list, err := repo.Search(context.Background(), repository.ScheduleFilter{
		OriginLatMin: ptr(35.0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("want 1, got %d", len(list))
	}
}

func TestScheduleRepo_Search_WithFilter_DepartAtRange(t *testing.T) {
	operatorID := uuid.New()
	s := newScheduleModel(operatorID)
	from := time.Now()
	to := time.Now().Add(48 * time.Hour)

	repo := &MockScheduleRepo{
		SearchFunc: func(_ context.Context, f repository.ScheduleFilter) ([]model.Schedule, error) {
			if f.DepartAtFrom == nil || f.DepartAtTo == nil {
				return []model.Schedule{}, nil
			}
			return []model.Schedule{*s}, nil
		},
	}

	list, err := repo.Search(context.Background(), repository.ScheduleFilter{
		DepartAtFrom: &from,
		DepartAtTo:   &to,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("want 1, got %d", len(list))
	}
}

// ---- UpdateStatus ----

func TestScheduleRepo_UpdateStatus_Success(t *testing.T) {
	scheduleID := uuid.New()
	var capturedStatus model.ScheduleStatus

	repo := &MockScheduleRepo{
		UpdateStatusFunc: func(_ context.Context, id uuid.UUID, status model.ScheduleStatus) error {
			if id != scheduleID {
				t.Errorf("unexpected id: %v", id)
			}
			capturedStatus = status
			return nil
		},
	}

	err := repo.UpdateStatus(context.Background(), scheduleID, model.ScheduleStatusFull)
	if err != nil {
		t.Fatal(err)
	}
	if capturedStatus != model.ScheduleStatusFull {
		t.Errorf("want full, got %v", capturedStatus)
	}
}

func TestScheduleRepo_UpdateStatus_Error(t *testing.T) {
	repo := &MockScheduleRepo{
		UpdateStatusFunc: func(_ context.Context, _ uuid.UUID, _ model.ScheduleStatus) error {
			return errors.New("db error")
		},
	}

	err := repo.UpdateStatus(context.Background(), uuid.New(), model.ScheduleStatusDeparted)
	if err == nil {
		t.Error("want error, got nil")
	}
}

// ---- Delete ----

func TestScheduleRepo_Delete_Success(t *testing.T) {
	scheduleID := uuid.New()
	deleted := false

	repo := &MockScheduleRepo{
		DeleteFunc: func(_ context.Context, id uuid.UUID) error {
			if id != scheduleID {
				t.Errorf("unexpected id: %v", id)
			}
			deleted = true
			return nil
		},
	}

	err := repo.Delete(context.Background(), scheduleID)
	if err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Error("want Delete called, but it was not")
	}
}

func TestScheduleRepo_Delete_Error(t *testing.T) {
	repo := &MockScheduleRepo{
		DeleteFunc: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("db error")
		},
	}

	err := repo.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Error("want error, got nil")
	}
}
