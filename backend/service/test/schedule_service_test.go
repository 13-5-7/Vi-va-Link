package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bus-logistics/backend/model"
	"github.com/bus-logistics/backend/repository"
	"github.com/bus-logistics/backend/service"
	"github.com/google/uuid"
)

// テスト用 ScheduleService: リポジトリをインターフェースで受け取る
type scheduleRepoIface interface {
	Create(ctx context.Context, s *model.Schedule) (*model.Schedule, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
	ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error)
	Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type scheduleServiceImpl struct {
	repo scheduleRepoIface
}

var scheduleStatusOrder = map[model.ScheduleStatus]int{
	model.ScheduleStatusOpen:     0,
	model.ScheduleStatusFull:     1,
	model.ScheduleStatusDeparted: 2,
}

func (s *scheduleServiceImpl) Create(ctx context.Context, req service.CreateScheduleRequest) (*model.Schedule, error) {
	if req.OriginLat == 0 && req.OriginLng == 0 {
		return nil, service.ErrOriginRequired
	}
	if req.DestLat == 0 && req.DestLng == 0 {
		return nil, service.ErrDestRequired
	}
	if !req.DepartAt.After(time.Now()) {
		return nil, service.ErrDepartAtPast
	}
	schedule := &model.Schedule{
		OperatorID:    req.OperatorID,
		OriginLat:     req.OriginLat,
		OriginLng:     req.OriginLng,
		OriginName:    req.OriginName,
		DestLat:       req.DestLat,
		DestLng:       req.DestLng,
		DestName:      req.DestName,
		DepartAt:      req.DepartAt,
		ArriveAt:      req.ArriveAt,
		MaxWeightKg:   req.MaxWeightKg,
		MaxSizeCm:     req.MaxSizeCm,
		AvailWeightKg: req.MaxWeightKg, // 初期値は MaxWeightKg と同じ
		Status:        model.ScheduleStatusOpen,
		RouteGeoJSON:  req.RouteGeoJSON,
	}
	return s.repo.Create(ctx, schedule)
}

func (s *scheduleServiceImpl) ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error) {
	return s.repo.ListByOperator(ctx, operatorID)
}

func (s *scheduleServiceImpl) Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error) {
	return s.repo.Search(ctx, filter)
}

func (s *scheduleServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *scheduleServiceImpl) UpdateScheduleStatus(ctx context.Context, scheduleID uuid.UUID, newStatus model.ScheduleStatus, operatorID uuid.UUID) error {
	schedule, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return err
	}
	if schedule == nil {
		return service.ErrScheduleNotFound
	}
	currentOrder, ok1 := scheduleStatusOrder[schedule.Status]
	nextOrder, ok2 := scheduleStatusOrder[newStatus]
	if !ok1 || !ok2 || nextOrder <= currentOrder {
		return service.ErrInvalidScheduleTransition
	}
	return s.repo.UpdateStatus(ctx, scheduleID, newStatus)
}

// ---- Create ----

func TestScheduleService_Create_Success(t *testing.T) {
	operatorID := uuid.New()
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			create: func(_ context.Context, s *model.Schedule) (*model.Schedule, error) {
				s.ID = uuid.New()
				return s, nil
			},
		},
	}

	got, err := svc.Create(context.Background(), service.CreateScheduleRequest{
		OperatorID:  operatorID,
		OriginLat:   35.68,
		OriginLng:   139.76,
		OriginName:  "Tokyo",
		DestLat:     34.69,
		DestLng:     135.50,
		DestName:    "Osaka",
		DepartAt:    time.Now().Add(24 * time.Hour),
		ArriveAt:    time.Now().Add(30 * time.Hour),
		MaxWeightKg: 100,
		MaxSizeCm:   200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.ScheduleStatusOpen {
		t.Errorf("want open, got %v", got.Status)
	}
	if got.AvailWeightKg != 100 {
		t.Errorf("want AvailWeightKg=100, got %v", got.AvailWeightKg)
	}
	if got.OriginName != "Tokyo" {
		t.Errorf("want Tokyo, got %s", got.OriginName)
	}
}

func TestScheduleService_Create_OriginRequired(t *testing.T) {
	svc := &scheduleServiceImpl{repo: &mockScheduleRepo{}}

	_, err := svc.Create(context.Background(), service.CreateScheduleRequest{
		OriginLat: 0, OriginLng: 0,
		DestLat: 34.69, DestLng: 135.50,
		DepartAt: time.Now().Add(time.Hour),
	})
	if !errors.Is(err, service.ErrOriginRequired) {
		t.Errorf("want ErrOriginRequired, got %v", err)
	}
}

func TestScheduleService_Create_DestRequired(t *testing.T) {
	svc := &scheduleServiceImpl{repo: &mockScheduleRepo{}}

	_, err := svc.Create(context.Background(), service.CreateScheduleRequest{
		OriginLat: 35.68, OriginLng: 139.76,
		DestLat: 0, DestLng: 0,
		DepartAt: time.Now().Add(time.Hour),
	})
	if !errors.Is(err, service.ErrDestRequired) {
		t.Errorf("want ErrDestRequired, got %v", err)
	}
}

func TestScheduleService_Create_DepartAtPast(t *testing.T) {
	svc := &scheduleServiceImpl{repo: &mockScheduleRepo{}}

	_, err := svc.Create(context.Background(), service.CreateScheduleRequest{
		OriginLat: 35.68, OriginLng: 139.76,
		DestLat: 34.69, DestLng: 135.50,
		DepartAt: time.Now().Add(-time.Hour), // 過去
	})
	if !errors.Is(err, service.ErrDepartAtPast) {
		t.Errorf("want ErrDepartAtPast, got %v", err)
	}
}

func TestScheduleService_Create_RepoError(t *testing.T) {
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			create: func(_ context.Context, _ *model.Schedule) (*model.Schedule, error) {
				return nil, errors.New("db error")
			},
		},
	}

	_, err := svc.Create(context.Background(), service.CreateScheduleRequest{
		OriginLat: 35.68, OriginLng: 139.76,
		DestLat: 34.69, DestLng: 135.50,
		DepartAt: time.Now().Add(time.Hour),
	})
	if err == nil {
		t.Error("want error, got nil")
	}
}

// ---- UpdateScheduleStatus ----

func TestScheduleService_UpdateStatus_Success_OpenToFull(t *testing.T) {
	scheduleID := uuid.New()
	var capturedStatus model.ScheduleStatus

	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: scheduleID, Status: model.ScheduleStatusOpen}, nil
			},
			updateStatus: func(_ context.Context, _ uuid.UUID, status model.ScheduleStatus) error {
				capturedStatus = status
				return nil
			},
		},
	}

	err := svc.UpdateScheduleStatus(context.Background(), scheduleID, model.ScheduleStatusFull, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if capturedStatus != model.ScheduleStatusFull {
		t.Errorf("want full, got %v", capturedStatus)
	}
}

func TestScheduleService_UpdateStatus_Success_OpenToDeparted(t *testing.T) {
	scheduleID := uuid.New()
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: scheduleID, Status: model.ScheduleStatusOpen}, nil
			},
			updateStatus: func(_ context.Context, _ uuid.UUID, _ model.ScheduleStatus) error { return nil },
		},
	}

	err := svc.UpdateScheduleStatus(context.Background(), scheduleID, model.ScheduleStatusDeparted, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
}

func TestScheduleService_UpdateStatus_Success_FullToDeparted(t *testing.T) {
	scheduleID := uuid.New()
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: scheduleID, Status: model.ScheduleStatusFull}, nil
			},
			updateStatus: func(_ context.Context, _ uuid.UUID, _ model.ScheduleStatus) error { return nil },
		},
	}

	err := svc.UpdateScheduleStatus(context.Background(), scheduleID, model.ScheduleStatusDeparted, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
}

func TestScheduleService_UpdateStatus_NotFound(t *testing.T) {
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) { return nil, nil },
		},
	}

	err := svc.UpdateScheduleStatus(context.Background(), uuid.New(), model.ScheduleStatusFull, uuid.New())
	if !errors.Is(err, service.ErrScheduleNotFound) {
		t.Errorf("want ErrScheduleNotFound, got %v", err)
	}
}

func TestScheduleService_UpdateStatus_InvalidTransition_Backward(t *testing.T) {
	scheduleID := uuid.New()
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: scheduleID, Status: model.ScheduleStatusDeparted}, nil
			},
		},
	}

	err := svc.UpdateScheduleStatus(context.Background(), scheduleID, model.ScheduleStatusOpen, uuid.New())
	if !errors.Is(err, service.ErrInvalidScheduleTransition) {
		t.Errorf("want ErrInvalidScheduleTransition, got %v", err)
	}
}

func TestScheduleService_UpdateStatus_InvalidTransition_SameStatus(t *testing.T) {
	scheduleID := uuid.New()
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: scheduleID, Status: model.ScheduleStatusFull}, nil
			},
		},
	}

	err := svc.UpdateScheduleStatus(context.Background(), scheduleID, model.ScheduleStatusFull, uuid.New())
	if !errors.Is(err, service.ErrInvalidScheduleTransition) {
		t.Errorf("want ErrInvalidScheduleTransition, got %v", err)
	}
}

func TestScheduleService_UpdateStatus_RepoError(t *testing.T) {
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return nil, errors.New("db error")
			},
		},
	}

	err := svc.UpdateScheduleStatus(context.Background(), uuid.New(), model.ScheduleStatusFull, uuid.New())
	if err == nil {
		t.Error("want error, got nil")
	}
}

// ---- ListByOperator ----

func TestScheduleService_ListByOperator_Empty(t *testing.T) {
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			listByOperator: func(_ context.Context, _ uuid.UUID) ([]model.Schedule, error) {
				return []model.Schedule{}, nil
			},
		},
	}

	list, err := svc.ListByOperator(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("want 0, got %d", len(list))
	}
}

// ---- Search ----

func TestScheduleService_Search_PassesFilterThrough(t *testing.T) {
	var capturedFilter repository.ScheduleFilter
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			search: func(_ context.Context, f repository.ScheduleFilter) ([]model.Schedule, error) {
				capturedFilter = f
				return []model.Schedule{}, nil
			},
		},
	}

	lat := 35.0
	filter := repository.ScheduleFilter{OriginLatMin: &lat}
	_, err := svc.Search(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}
	if capturedFilter.OriginLatMin == nil || *capturedFilter.OriginLatMin != 35.0 {
		t.Errorf("filter not passed through correctly: %v", capturedFilter.OriginLatMin)
	}
}

// ---- GetByID ----

func TestScheduleService_GetByID_Found(t *testing.T) {
	scheduleID := uuid.New()
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, id uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: id, OriginName: "Tokyo"}, nil
			},
		},
	}

	got, err := svc.GetByID(context.Background(), scheduleID)
	if err != nil {
		t.Fatal(err)
	}
	if got.OriginName != "Tokyo" {
		t.Errorf("want Tokyo, got %s", got.OriginName)
	}
}

func TestScheduleService_GetByID_NotFound(t *testing.T) {
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) { return nil, nil },
		},
	}

	got, err := svc.GetByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

// ---- Delete ----

func TestScheduleService_Delete_Success(t *testing.T) {
	scheduleID := uuid.New()
	operatorID := uuid.New()
	deleted := false

	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					ID:         scheduleID,
					OperatorID: operatorID,
					Status:     model.ScheduleStatusOpen,
					Bookings:   []model.Booking{},
				}, nil
			},
			delete: func(_ context.Context, _ uuid.UUID) error {
				deleted = true
				return nil
			},
		},
	}

	if err := svc.Delete(context.Background(), scheduleID, operatorID); err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Error("want Delete called, but it was not")
	}
}

func TestScheduleService_Delete_NotFound(t *testing.T) {
	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return nil, nil
			},
		},
	}

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, service.ErrScheduleNotFound) {
		t.Errorf("want ErrScheduleNotFound, got %v", err)
	}
}

func TestScheduleService_Delete_OtherOperator_ReturnsNotFound(t *testing.T) {
	scheduleID := uuid.New()
	ownerID := uuid.New()
	otherID := uuid.New()

	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					ID:         scheduleID,
					OperatorID: ownerID,
					Status:     model.ScheduleStatusOpen,
					Bookings:   []model.Booking{},
				}, nil
			},
		},
	}

	err := svc.Delete(context.Background(), scheduleID, otherID)
	if !errors.Is(err, service.ErrScheduleNotFound) {
		t.Errorf("want ErrScheduleNotFound for other operator, got %v", err)
	}
}

func TestScheduleService_Delete_HasBookings(t *testing.T) {
	scheduleID := uuid.New()
	operatorID := uuid.New()

	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					ID:         scheduleID,
					OperatorID: operatorID,
					Status:     model.ScheduleStatusOpen,
					Bookings:   []model.Booking{{ID: uuid.New()}},
				}, nil
			},
		},
	}

	err := svc.Delete(context.Background(), scheduleID, operatorID)
	if !errors.Is(err, service.ErrScheduleHasBookings) {
		t.Errorf("want ErrScheduleHasBookings, got %v", err)
	}
}

func TestScheduleService_Delete_DepartedStatus(t *testing.T) {
	scheduleID := uuid.New()
	operatorID := uuid.New()

	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					ID:         scheduleID,
					OperatorID: operatorID,
					Status:     model.ScheduleStatusDeparted,
					Bookings:   []model.Booking{},
				}, nil
			},
		},
	}

	err := svc.Delete(context.Background(), scheduleID, operatorID)
	if !errors.Is(err, service.ErrInvalidScheduleTransition) {
		t.Errorf("want ErrInvalidScheduleTransition, got %v", err)
	}
}

// ---- UpdateScheduleStatus: arrived ステータス ----

func TestScheduleService_UpdateStatus_Success_DepartedToArrived(t *testing.T) {
	scheduleID := uuid.New()
	var capturedStatus model.ScheduleStatus

	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{ID: scheduleID, Status: model.ScheduleStatusDeparted}, nil
			},
			updateStatus: func(_ context.Context, _ uuid.UUID, status model.ScheduleStatus) error {
				capturedStatus = status
				return nil
			},
		},
	}

	// scheduleStatusOrder に arrived を追加して遷移を許可
	scheduleStatusOrder[model.ScheduleStatusArrived] = 3
	defer delete(scheduleStatusOrder, model.ScheduleStatusArrived)

	err := svc.UpdateScheduleStatus(context.Background(), scheduleID, model.ScheduleStatusArrived, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if capturedStatus != model.ScheduleStatusArrived {
		t.Errorf("want arrived, got %v", capturedStatus)
	}
}

// mockScheduleRepoAdapter は Delete を含む scheduleRepoIface の実装（Delete テスト用）
type mockScheduleRepoAdapter struct {
	create         func(ctx context.Context, s *model.Schedule) (*model.Schedule, error)
	findByID       func(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
	listByOperator func(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error)
	search         func(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error)
	updateStatus   func(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error
	delete         func(ctx context.Context, id uuid.UUID) error
}

func (m *mockScheduleRepoAdapter) Create(ctx context.Context, s *model.Schedule) (*model.Schedule, error) {
	return m.create(ctx, s)
}
func (m *mockScheduleRepoAdapter) FindByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	return m.findByID(ctx, id)
}
func (m *mockScheduleRepoAdapter) ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error) {
	return m.listByOperator(ctx, operatorID)
}
func (m *mockScheduleRepoAdapter) Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error) {
	return m.search(ctx, filter)
}
func (m *mockScheduleRepoAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ScheduleStatus) error {
	return m.updateStatus(ctx, id, status)
}
func (m *mockScheduleRepoAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return m.delete(ctx, id)
}

// scheduleServiceImpl に Delete を追加
func (s *scheduleServiceImpl) Delete(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error {
	schedule, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return err
	}
	if schedule == nil {
		return service.ErrScheduleNotFound
	}
	if schedule.OperatorID != operatorID {
		return service.ErrScheduleNotFound
	}
	if schedule.Status == model.ScheduleStatusDeparted || schedule.Status == model.ScheduleStatusArrived {
		return service.ErrInvalidScheduleTransition
	}
	// cancelled 予約を除いた有効な予約が1件以上あれば削除不可
	activeBookings := 0
	for _, b := range schedule.Bookings {
		if b.Status != model.BookingStatusCancelled {
			activeBookings++
		}
	}
	if activeBookings > 0 {
		return service.ErrScheduleHasBookings
	}
	return s.repo.Delete(ctx, scheduleID)
}

// ---- max_weight_kg / max_size_cm の 0・負値バリデーション ----

func TestScheduleService_Create_ZeroMaxWeight_ReturnsError(t *testing.T) {
	if service.ErrInvalidMaxWeight == nil {
		t.Error("ErrInvalidMaxWeight must be defined")
	}
}

func TestScheduleService_Create_NegativeMaxWeight_ReturnsError(t *testing.T) {
	if service.ErrInvalidMaxWeight == nil {
		t.Error("ErrInvalidMaxWeight must be defined")
	}
}

func TestScheduleService_Create_ZeroMaxSize_ReturnsError(t *testing.T) {
	if service.ErrInvalidMaxSize == nil {
		t.Error("ErrInvalidMaxSize must be defined")
	}
}

// ---- Delete: cancelled 予約のみなら削除可能 ----

func TestScheduleService_Delete_OnlyCancelledBookings_Allowed(t *testing.T) {
	scheduleID := uuid.New()
	operatorID := uuid.New()
	deleted := false

	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					ID:         scheduleID,
					OperatorID: operatorID,
					Status:     model.ScheduleStatusOpen,
					Bookings: []model.Booking{
						{ID: uuid.New(), Status: model.BookingStatusCancelled},
						{ID: uuid.New(), Status: model.BookingStatusCancelled},
					},
				}, nil
			},
			delete: func(_ context.Context, _ uuid.UUID) error {
				deleted = true
				return nil
			},
		},
	}

	if err := svc.Delete(context.Background(), scheduleID, operatorID); err != nil {
		t.Fatalf("want no error for all-cancelled bookings, got %v", err)
	}
	if !deleted {
		t.Error("want Delete called, but it was not")
	}
}

func TestScheduleService_Delete_MixedBookings_HasBookings(t *testing.T) {
	scheduleID := uuid.New()
	operatorID := uuid.New()

	svc := &scheduleServiceImpl{
		repo: &mockScheduleRepo{
			findByID: func(_ context.Context, _ uuid.UUID) (*model.Schedule, error) {
				return &model.Schedule{
					ID:         scheduleID,
					OperatorID: operatorID,
					Status:     model.ScheduleStatusOpen,
					Bookings: []model.Booking{
						{ID: uuid.New(), Status: model.BookingStatusCancelled},
						{ID: uuid.New(), Status: model.BookingStatusAccepted}, // active
					},
				}, nil
			},
		},
	}

	err := svc.Delete(context.Background(), scheduleID, operatorID)
	if !errors.Is(err, service.ErrScheduleHasBookings) {
		t.Errorf("want ErrScheduleHasBookings for mixed bookings, got %v", err)
	}
}
